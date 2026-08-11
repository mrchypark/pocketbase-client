import json
import os
import sys
import urllib.error
import urllib.request
import time
import argparse

# Constants
AUTODATE_CREATED_FIELD = {
    "hidden": False,
    "id": "autodate_created",
    "name": "created",
    "onCreate": True,
    "onUpdate": False,
    "presentable": False,
    "system": True,
    "type": "autodate"
}

AUTODATE_UPDATED_FIELD = {
    "hidden": False,
    "id": "autodate_updated",
    "name": "updated",
    "onCreate": True,
    "onUpdate": True,
    "presentable": False,
    "system": True,
    "type": "autodate"
}

def request_json(base_url, method, path, token=None, payload=None):
    url = base_url.rstrip("/") + path
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = token
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")

    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            if response.status == 204:
                return {}
            body = response.read().decode("utf-8")
            if not body:
                return {}
            return json.loads(body)
    except urllib.error.HTTPError as err:
        detail = err.read().decode("utf-8")
        print(f"HTTP Error {err.code}: {detail}", file=sys.stderr)
        raise RuntimeError(f"{method} {path} failed: {err.code} {detail}") from err
    except urllib.error.URLError as err:
        raise RuntimeError(f"{method} {path} failed: {err}") from err

def get_auth_token(base_url, email, password):
    auth_paths = [
        "/api/collections/_superusers/auth-with-password",
        "/api/admins/auth-with-password",
    ]

    # Wait for server
    for i in range(5):
        try:
            request_json(base_url, "GET", "/api/health")
            break
        except Exception:
            print(f"Waiting for server... ({i + 1}/5)")
            time.sleep(2)

    for path in auth_paths:
        try:
            auth = request_json(
                base_url, "POST", path,
                payload={"identity": email, "password": password}
            )
            return auth.get("token")
        except RuntimeError:
            continue
    return None

def add_autodate_fields(col_def):
    if col_def.get("type") == "view":
        return

    existing_fields = col_def.get("fields", [])
    existing_names = {f.get("name") for f in existing_fields}

    if "created" not in existing_names:
        col_def["fields"] = [AUTODATE_CREATED_FIELD] + existing_fields

    if "updated" not in existing_names:
        current_names = {f.get("name") for f in col_def["fields"]}
        if "updated" not in current_names:
            col_def["fields"] = col_def["fields"] + [AUTODATE_UPDATED_FIELD]

def resolve_collection_ids(col_def, name_to_id):
    """Resolve collection names to IDs in relation fields."""
    for field in col_def.get("fields", []):
        if "options" in field and isinstance(field["options"], dict):
            for k, v in field["options"].items():
                field[k] = v

        target = field.get("collectionId")
        if not target and "options" in field:
            target = field["options"].get("collectionId")

        if target and target in name_to_id:
            real_id = name_to_id[target]
            field["collectionId"] = real_id
            if "options" in field:
                field["options"]["collectionId"] = real_id

def cmd_apply(base_url, token, schema_path):
    if not os.path.exists(schema_path):
        print(f"Schema file not found: {schema_path}", file=sys.stderr)
        sys.exit(1)

    with open(schema_path, "r") as f:
        target_collections = json.load(f)

    existing_res = request_json(base_url, "GET", "/api/collections?perPage=200", token=token)
    items = existing_res.get("items", existing_res if isinstance(existing_res, list) else [])
    existing_map = {item["name"]: item for item in items}
    name_to_id = {item["name"]: item["id"] for item in items}

    for col_def in target_collections:
        name = col_def["name"]
        if name not in name_to_id:
            payload = col_def.copy()
            add_autodate_fields(payload)
            resolve_collection_ids(payload, name_to_id)
            try:
                print(f"Creating collection: {name}")
                created = request_json(base_url, "POST", "/api/collections", token=token, payload=payload)
                name_to_id[name] = created["id"]
                existing_map[name] = created
            except Exception as e:
                print(f"Creation failed for {name}: {e}. Will retry in update pass.")

    for col_def in target_collections:
        name = col_def["name"]
        if name not in existing_map:
            continue

        item_id = existing_map[name]["id"]

        if col_def.get("type") == "view":
            pass

        payload = col_def.copy()
        add_autodate_fields(payload)
        resolve_collection_ids(payload, name_to_id)

        payload["id"] = item_id

        try:
            print(f"Updating collection: {name}")
            request_json(base_url, "PATCH", f"/api/collections/{item_id}", token=token, payload=payload)
        except Exception as e:
            print(f"Update failed for {name}: {e}")

    print("Schema applied successfully.")

def cmd_dump(base_url, token, schema_path):
    try:
        res = request_json(base_url, "GET", "/api/collections?perPage=200", token=token)
        all_items = res.get("items", [])

        items = [item for item in all_items if not item["name"].startswith("_")]
        items.sort(key=lambda x: x["name"])

        with open(schema_path, "w") as f:
            json.dump(items, f, indent=4)
        print(f"Schema dumped to {schema_path}")
    except Exception as e:
        print(f"Dump failed: {e}", file=sys.stderr)
        sys.exit(1)

def cmd_create_collection(base_url, token, name, col_type):
    payload = {
        "name": name,
        "type": col_type,
        "fields": []
    }
    try:
        request_json(base_url, "POST", "/api/collections", token=token, payload=payload)
        print(f"Collection '{name}' created.")
    except Exception as e:
        print(f"Failed to create collection: {e}", file=sys.stderr)
        sys.exit(1)

def cmd_add_field(base_url, token, collection, field_json):
    try:
        field_def = json.loads(field_json)
    except json.JSONDecodeError as e:
        print(f"Invalid JSON for field definition: {e}", file=sys.stderr)
        sys.exit(1)

    col_id = get_collection_id(base_url, token, collection)
    if not col_id:
        print(f"Collection '{collection}' not found.", file=sys.stderr)
        sys.exit(1)

    try:
        col = request_json(base_url, "GET", f"/api/collections/{col_id}", token=token)
        existing_fields = col.get("fields", [])

        for f in existing_fields:
            if f["name"] == field_def["name"]:
                print(f"Field '{field_def['name']}' already exists. Updating...")
                f.update(field_def)
                break
        else:
            existing_fields.append(field_def)

        payload = {"fields": existing_fields}
        request_json(base_url, "PATCH", f"/api/collections/{col_id}", token=token, payload=payload)
        print(f"Field added/updated in '{collection}'.")
    except Exception as e:
        print(f"Failed to add field: {e}", file=sys.stderr)
        sys.exit(1)

def cmd_delete_field(base_url, token, collection, field_name):
    col_id = get_collection_id(base_url, token, collection)
    if not col_id:
        print(f"Collection '{collection}' not found.", file=sys.stderr)
        sys.exit(1)

    try:
        col = request_json(base_url, "GET", f"/api/collections/{col_id}", token=token)
        original_fields = col.get("fields", [])
        new_fields = [f for f in original_fields if f["name"] != field_name]

        if len(original_fields) == len(new_fields):
            print(f"Field '{field_name}' not found in '{collection}'.")
            return

        payload = {"fields": new_fields}
        request_json(base_url, "PATCH", f"/api/collections/{col_id}", token=token, payload=payload)
        print(f"Field '{field_name}' deleted from '{collection}'.")
    except Exception as e:
        print(f"Failed to delete field: {e}", file=sys.stderr)
        sys.exit(1)

def get_collection_id(base_url, token, name_or_id):
    try:
        res = request_json(base_url, "GET", f"/api/collections/{name_or_id}", token=token)
        return res["id"]
    except:
        try:
            res = request_json(base_url, "GET", f"/api/collections?filter=(name='{name_or_id}')", token=token)
            items = res.get("items", [])
            if items:
                return items[0]["id"]
        except:
            pass
    return None

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=["apply", "dump", "create_collection", "add_field", "delete_field"], help="Action to perform")
    parser.add_argument("--url", default=os.environ.get("PB_URL", "http://127.0.0.1:8090"))
    parser.add_argument("--email", default=os.environ.get("PB_ADMIN_EMAIL"))
    parser.add_argument("--password", default=os.environ.get("PB_ADMIN_PASSWORD"))
    parser.add_argument("--schema", default="pb_schema.json")

    parser.add_argument("--name", help="Collection name for create")
    parser.add_argument("--type", default="base", help="Collection type (base, auth, view)")
    parser.add_argument("--collection", help="Target collection for field ops")
    parser.add_argument("--field-json", help="JSON definition of field to add")
    parser.add_argument("--field-name", help="Name of field to delete")

    args = parser.parse_args()

    if not args.email or not args.password:
        print("Missing credentials (PB_ADMIN_EMAIL/PASSWORD)", file=sys.stderr)
        sys.exit(1)

    token = get_auth_token(args.url, args.email, args.password)
    if not token:
        print("Authentication failed", file=sys.stderr)
        sys.exit(1)

    if args.command == "apply":
        cmd_apply(args.url, token, args.schema)
    elif args.command == "dump":
        cmd_dump(args.url, token, args.schema)
    elif args.command == "create_collection":
        cmd_create_collection(args.url, token, args.name, args.type)
    elif args.command == "add_field":
        cmd_add_field(args.url, token, args.collection, args.field_json)
    elif args.command == "delete_field":
        cmd_delete_field(args.url, token, args.collection, args.field_name)

if __name__ == "__main__":
    main()
