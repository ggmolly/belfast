import os
import re

REGIONS = ["EN", "CN", "JP", "KR", "TW"]
REGION_ORDER = {"CN": 0, "JP": 1, "KR": 2, "TW": 3}

LABEL_MAP = {
    1: "optional",
    2: "required",
    3: "repeated",
}

TYPE_MAP = {
    1: "double",
    2: "float",
    3: "int64",
    4: "uint64",
    5: "int32",
    6: "fixed64",
    7: "fixed32",
    8: "bool",
    9: "string",
    10: "group",
    11: "message",
    12: "bytes",
    13: "uint32",
    14: "enum",
    15: "sfixed32",
    16: "sfixed64",
    17: "sint32",
    18: "sint64",
}

DESCRIPTOR_RE = re.compile(r"^(\w+)\s*=\s*slot\d+\.Descriptor\(\)")
FIELD_DESC_RE = re.compile(
    r"^(slot\d+\.[A-Za-z0-9_]+_FIELD_LIST\.[A-Za-z0-9_]+)\s*=\s*slot\d+\.FieldDescriptor\(\)"
)
PROPERTY_RE = re.compile(r"^([A-Za-z0-9_\.]+)\.([A-Za-z0-9_]+)\s*=\s*(.+)$")
FIELDS_START_RE = re.compile(r"^(\w+)\.fields\s*=\s*{")
NAMESPACE_FILE_RE = re.compile(r"^(p\d+)_pb\.lua$")


class FieldInfo:
    def __init__(self, symbol, source_file, namespace):
        self.symbol = symbol
        self.source_file = source_file
        self.namespace = namespace
        self.name = ""
        self.full_name = ""
        self.number = 0
        self.index = 0
        self.label = 0
        self.type = 0
        self.cpp_type = 0
        self.message_type_symbol = ""
        self.enum_type_symbol = ""

    def signature(self):
        return (
            self.name,
            self.number,
            self.label,
            self.type,
            self.message_type_symbol,
            self.enum_type_symbol,
            self.namespace,
        )


class MessageInfo:
    def __init__(self, symbol, source_file):
        self.symbol = symbol
        self.source_file = source_file
        self.name = ""
        self.full_name = ""
        self.namespace = ""
        self.field_symbols = []

    def key(self):
        return (self.symbol, self.namespace)

    def field_signature_list(self, field_map):
        fields = [field_map[key] for key in self.field_symbols]
        fields.sort(key=lambda item: item.index)
        return [field.signature() for field in fields]


class RegionData:
    def __init__(self, region):
        self.region = region
        self.messages = {}
        self.fields = {}
        self.symbol_namespaces = {}


def log_progress(message):
    print(f"\r\033[2K{message}", end="", flush=True)


def log_line(message):
    print(f"\r\033[2K{message}", flush=True)


def log_final(message):
    print(f"\033[2K{message}", flush=True)


def parse_string(value):
    value = value.strip()
    if value.startswith('"') and value.endswith('"'):
        return value[1:-1]
    return value


def parse_symbol(value):
    value = value.strip()
    return value.split(".")[-1]


def parse_field_list(lines, start_index, namespace):
    items = []
    i = start_index
    while i < len(lines):
        line = lines[i].strip()
        match = re.search(r"(slot\d+\.[A-Za-z0-9_]+_FIELD_LIST\.[A-Za-z0-9_]+)", line)
        if match:
            items.append((match.group(1), namespace))
        if "}" in line:
            return items, i
        i += 1
    return items, i


def parse_lua_file(path, region_data):
    with open(path, "r") as file:
        lines = file.readlines()

    namespace = ""
    match = NAMESPACE_FILE_RE.match(os.path.basename(path))
    if match:
        namespace = match.group(1)

    i = 0
    while i < len(lines):
        line = lines[i].strip()
        if not line:
            i += 1
            continue
        descriptor_match = DESCRIPTOR_RE.match(line)
        if descriptor_match:
            symbol = descriptor_match.group(1)
            key = (symbol, namespace)
            message = MessageInfo(symbol, path)
            message.namespace = namespace
            region_data.messages[key] = message
            region_data.symbol_namespaces.setdefault(symbol, set()).add(namespace)
            i += 1
            continue
        field_desc_match = FIELD_DESC_RE.match(line)
        if field_desc_match:
            symbol = field_desc_match.group(1)
            key = (symbol, namespace)
            region_data.fields[key] = FieldInfo(symbol, path, namespace)
            i += 1
            continue
        fields_start_match = FIELDS_START_RE.match(line)
        if fields_start_match:
            symbol = fields_start_match.group(1)
            items, end_index = parse_field_list(lines, i + 1, namespace)
            key = (symbol, namespace)
            if key in region_data.messages:
                region_data.messages[key].field_symbols = items
            i = end_index + 1
            continue
        property_match = PROPERTY_RE.match(line)
        if property_match:
            target = property_match.group(1)
            prop = property_match.group(2)
            value = property_match.group(3).strip()
            field_key = (target, namespace)
            if field_key in region_data.fields:
                field = region_data.fields[field_key]
                if prop == "name":
                    field.name = parse_string(value)
                elif prop == "full_name":
                    field.full_name = parse_string(value)
                elif prop == "number":
                    field.number = int(value)
                elif prop == "index":
                    field.index = int(value)
                elif prop == "label":
                    field.label = int(value)
                elif prop == "type":
                    field.type = int(value)
                elif prop == "cpp_type":
                    field.cpp_type = int(value)
                elif prop == "message_type":
                    field.message_type_symbol = parse_symbol(value)
                elif prop == "enum_type":
                    field.enum_type_symbol = parse_symbol(value)
            else:
                target_match = re.match(r"^(\w+)\.", target)
                if target_match:
                    symbol = target_match.group(1)
                    key = (symbol, namespace)
                    if key in region_data.messages:
                        message = region_data.messages[key]
                        if prop == "name":
                            message.name = parse_string(value)
                        elif prop == "full_name":
                            message.full_name = parse_string(value)
            i += 1
            continue
        i += 1


def parse_region(region, repo_root):
    region_data = RegionData(region)
    region_path = os.path.join(
        repo_root, "AzurLaneLuaScripts", region, "net", "protocol"
    )
    files = [
        os.path.join(region_path, name)
        for name in os.listdir(region_path)
        if name.endswith(".lua")
    ]
    files.sort()
    for path in files:
        rel_path = os.path.relpath(path, repo_root)
        log_progress(f"{region} parsing {rel_path}")
        parse_lua_file(path, region_data)
        log_line(f"{region} parsed {rel_path} ok")
    return region_data


def build_signatures(region_data):
    signatures = {}
    for key, message in region_data.messages.items():
        signatures[key] = message.field_signature_list(region_data.fields)
    return signatures


def format_namespace_suffix(namespace):
    if not namespace:
        return ""
    return namespace.upper()


def resolve_message_key(symbol, region_data, namespace):
    candidate = (symbol, namespace)
    if candidate in region_data.messages:
        return candidate
    namespaces = region_data.symbol_namespaces.get(symbol)
    if not namespaces:
        return candidate
    if len(namespaces) == 1:
        return (symbol, next(iter(namespaces)))
    if namespace:
        return (symbol, namespace)
    return (symbol, sorted(namespaces)[0])


def group_variants(en_signatures, region_signatures):
    variant_suffix_by_message_region = {}
    variant_groups = {}
    for key, en_signature in en_signatures.items():
        variant_groups[key] = {}
        for region, signatures in region_signatures.items():
            signature = signatures.get(key)
            if signature is None:
                continue
            if signature != en_signature:
                variant_groups[key].setdefault(tuple(signature), []).append(region)
        for signature, regions in variant_groups[key].items():
            regions.sort(key=lambda item: REGION_ORDER[item])
            suffix = "_".join(regions)
            for region in regions:
                variant_suffix_by_message_region.setdefault(key, {})[region] = suffix
    return variant_suffix_by_message_region, variant_groups


def build_message_name_map(en_data):
    name_by_key = {}
    for symbol, namespaces in en_data.symbol_namespaces.items():
        suffix_needed = len(namespaces) > 1
        for namespace in sorted(namespaces):
            key = (symbol, namespace)
            message = en_data.messages.get(key)
            if message is None:
                continue
            base_name = message.name.upper() if message.name else symbol
            if suffix_needed and namespace:
                namespace_suffix = format_namespace_suffix(namespace)
                base_name = f"{base_name}_{namespace_suffix}"
            name_by_key[key] = base_name
        if len(namespaces) == 1:
            key = (symbol, "")
            message = en_data.messages.get(key)
            if message is None:
                continue
            base_name = message.name.upper() if message.name else symbol
            name_by_key[key] = base_name
    return name_by_key


def resolve_type_name(
    field,
    region,
    variant_suffix_by_message_region,
    name_by_key,
    region_data,
    namespace,
):
    if field.type == 11:
        message_key = resolve_message_key(
            field.message_type_symbol, region_data, namespace
        )
        message_name = name_by_key.get(message_key, field.message_type_symbol)
        suffix = variant_suffix_by_message_region.get(message_key, {}).get(region, "")
        if suffix:
            return f"{message_name}_{suffix}"
        return message_name
    if field.type == 14:
        enum_key = resolve_message_key(field.enum_type_symbol, region_data, namespace)
        return name_by_key.get(enum_key, field.enum_type_symbol)
    return TYPE_MAP[field.type]


def resolve_file_name(
    message_key, region, variant_suffix_by_message_region, name_by_key
):
    message_name = name_by_key.get(message_key, message_key[0])
    suffix = variant_suffix_by_message_region.get(message_key, {}).get(region, "")
    if suffix:
        return f"{message_name}_{suffix}.proto"
    return f"{message_name}.proto"


def render_proto(
    message_key,
    message_name,
    message,
    region,
    variant_suffix_by_message_region,
    field_map,
    suffix,
    name_by_key,
    region_data,
):
    rendered_message_name = f"{message_name}_{suffix}" if suffix else message_name
    fields = [field_map[key] for key in message.field_symbols]
    fields.sort(key=lambda item: item.index)
    imports = set()
    for field in fields:
        if field.type == 11:
            target_key = resolve_message_key(
                field.message_type_symbol, region_data, message.namespace
            )
            target_name = resolve_file_name(
                target_key, region, variant_suffix_by_message_region, name_by_key
            )
            if target_name != f"{rendered_message_name}.proto":
                imports.add(target_name)
        elif field.type == 14:
            enum_key = resolve_message_key(
                field.enum_type_symbol, region_data, message.namespace
            )
            target_name = resolve_file_name(
                enum_key,
                region,
                variant_suffix_by_message_region,
                name_by_key,
            )
            if target_name != f"{rendered_message_name}.proto":
                imports.add(target_name)
    lines = [
        'syntax = "proto2";',
        "",
        "package belfast;",
        "",
        'option go_package = "./protobuf";',
        "",
    ]
    if imports:
        for name in sorted(imports):
            lines.append(f'import "{name}";')
        lines.append("")
    lines.append(f"message {rendered_message_name} {{")
    seen_numbers = set()
    for field in fields:
        if field.number in seen_numbers:
            continue
        seen_numbers.add(field.number)
        label = LABEL_MAP[field.label]
        field_type = resolve_type_name(
            field,
            region,
            variant_suffix_by_message_region,
            name_by_key,
            region_data,
            message.namespace,
        )
        lines.append(f"  {label} {field_type} {field.name} = {field.number};")
    lines.append("}")
    lines.append("")
    return "\n".join(lines)


def generate_outputs(
    repo_root,
    en_data,
    region_data_map,
    variant_suffix_by_message_region,
    variant_groups,
    name_by_key,
):
    output_dir = os.path.join(repo_root, "internal", "proto")
    os.makedirs(output_dir, exist_ok=True)
    outputs = []
    for key, message in en_data.messages.items():
        message_name = name_by_key.get(key, key[0])
        outputs.append((key, message_name, message, "EN", ""))
    for symbol, namespaces in en_data.symbol_namespaces.items():
        if len(namespaces) <= 1:
            continue
        default_key = (symbol, "")
        if default_key in en_data.messages:
            continue
        namespace = sorted(namespaces)[0]
        fallback_key = (symbol, namespace)
        if fallback_key not in en_data.messages:
            continue
        message = en_data.messages[fallback_key]
        message_name = name_by_key.get(fallback_key, symbol)
        outputs.append((default_key, message_name, message, "EN", ""))
    for key, signature_groups in variant_groups.items():
        for signature, regions in signature_groups.items():
            region = regions[0]
            suffix = variant_suffix_by_message_region[key][region]
            message = region_data_map[region].messages.get(key)
            message_name = name_by_key.get(key, key[0])
            outputs.append((key, message_name, message, region, suffix))
    outputs.sort(key=lambda item: (item[3], item[1], item[4]))
    count = 0
    for key, message_name, message, region, suffix in outputs:
        file_name = (
            f"{message_name}_{suffix}.proto" if suffix else f"{message_name}.proto"
        )
        target_path = os.path.join(output_dir, file_name)
        log_progress(f"writing {os.path.relpath(target_path, repo_root)}")
        content = render_proto(
            key,
            message_name,
            message,
            region,
            variant_suffix_by_message_region,
            region_data_map[region].fields,
            suffix,
            name_by_key,
            region_data_map[region],
        )
        with open(target_path, "w") as file:
            file.write(content)
        count += 1
    log_final(f"wrote {count} proto files to internal/proto")


def main():
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
    region_data_map = {}
    for region in REGIONS:
        region_data_map[region] = parse_region(region, repo_root)
    en_data = region_data_map["EN"]
    en_signatures = build_signatures(en_data)
    other_signatures = {}
    for region in REGIONS:
        if region == "EN":
            continue
        other_signatures[region] = build_signatures(region_data_map[region])
    name_by_key = build_message_name_map(en_data)
    variant_suffix_by_message_region, variant_groups = group_variants(
        en_signatures, other_signatures
    )
    generate_outputs(
        repo_root,
        en_data,
        region_data_map,
        variant_suffix_by_message_region,
        variant_groups,
        name_by_key,
    )


if __name__ == "__main__":
    main()
