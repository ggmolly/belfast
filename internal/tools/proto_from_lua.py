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


class FieldInfo:
    def __init__(self, symbol, source_file):
        self.symbol = symbol
        self.source_file = source_file
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
        )


class MessageInfo:
    def __init__(self, symbol, source_file):
        self.symbol = symbol
        self.source_file = source_file
        self.name = ""
        self.full_name = ""
        self.field_symbols = []

    def field_signature_list(self, field_map):
        fields = [field_map[symbol] for symbol in self.field_symbols]
        fields.sort(key=lambda item: item.index)
        return [field.signature() for field in fields]


class RegionData:
    def __init__(self, region):
        self.region = region
        self.messages = {}
        self.fields = {}


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


def parse_field_list(lines, start_index):
    items = []
    i = start_index
    while i < len(lines):
        line = lines[i].strip()
        match = re.search(r"(slot\d+\.[A-Za-z0-9_]+_FIELD_LIST\.[A-Za-z0-9_]+)", line)
        if match:
            items.append(match.group(1))
        if "}" in line:
            return items, i
        i += 1
    return items, i


def parse_lua_file(path, region_data):
    with open(path, "r") as file:
        lines = file.readlines()

    i = 0
    while i < len(lines):
        line = lines[i].strip()
        if not line:
            i += 1
            continue
        descriptor_match = DESCRIPTOR_RE.match(line)
        if descriptor_match:
            symbol = descriptor_match.group(1)
            region_data.messages[symbol] = MessageInfo(symbol, path)
            i += 1
            continue
        field_desc_match = FIELD_DESC_RE.match(line)
        if field_desc_match:
            symbol = field_desc_match.group(1)
            region_data.fields[symbol] = FieldInfo(symbol, path)
            i += 1
            continue
        fields_start_match = FIELDS_START_RE.match(line)
        if fields_start_match:
            symbol = fields_start_match.group(1)
            items, end_index = parse_field_list(lines, i + 1)
            if symbol in region_data.messages:
                region_data.messages[symbol].field_symbols = items
            i = end_index + 1
            continue
        property_match = PROPERTY_RE.match(line)
        if property_match:
            target = property_match.group(1)
            prop = property_match.group(2)
            value = property_match.group(3).strip()
            if target in region_data.fields:
                field = region_data.fields[target]
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
            elif target in region_data.messages:
                message = region_data.messages[target]
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
    for symbol, message in region_data.messages.items():
        signatures[symbol] = message.field_signature_list(region_data.fields)
    return signatures


def resolve_symbol(symbol, symbol_to_normalized):
    return symbol_to_normalized.get(symbol, symbol)


def group_variants(en_signatures, region_signatures, symbol_to_normalized):
    variant_suffix_by_message_region = {}
    variant_groups = {}
    for symbol, en_signature in en_signatures.items():
        normalized_symbol = resolve_symbol(symbol, symbol_to_normalized)
        variant_groups[normalized_symbol] = {}
        for region, signatures in region_signatures.items():
            signature = signatures.get(symbol)
            if signature is None:
                continue
            if signature != en_signature:
                variant_groups[normalized_symbol].setdefault(
                    tuple(signature), []
                ).append(region)
        for signature, regions in variant_groups[normalized_symbol].items():
            regions.sort(key=lambda item: REGION_ORDER[item])
            suffix = "_".join(regions)
            for region in regions:
                variant_suffix_by_message_region.setdefault(symbol, {})[region] = suffix
    return variant_suffix_by_message_region, variant_groups


def resolve_type_name(field, region, variant_suffix_by_message_region, symbol_to_name):
    if field.type == 11:
        message_symbol = resolve_symbol(field.message_type_symbol, symbol_to_name)
        suffix = variant_suffix_by_message_region.get(message_symbol, {}).get(
            region, ""
        )
        if suffix:
            return f"{message_symbol}_{suffix}"
        return message_symbol
    if field.type == 14:
        return resolve_symbol(field.enum_type_symbol, symbol_to_name)
    return TYPE_MAP[field.type]


def resolve_file_name(message_symbol, region, variant_suffix_by_message_region):
    suffix = variant_suffix_by_message_region.get(message_symbol, {}).get(region, "")
    if suffix:
        return f"{message_symbol}_{suffix}.proto"
    return f"{message_symbol}.proto"


def render_proto(
    message_symbol,
    message,
    region,
    variant_suffix_by_message_region,
    field_map,
    suffix,
    symbol_to_name,
):
    message_name = f"{message_symbol}_{suffix}" if suffix else message_symbol
    fields = [field_map[symbol] for symbol in message.field_symbols]
    fields.sort(key=lambda item: item.index)
    imports = set()
    for field in fields:
        if field.type == 11:
            field_symbol = resolve_symbol(field.message_type_symbol, symbol_to_name)
            target_name = resolve_file_name(
                field_symbol, region, variant_suffix_by_message_region
            )
            if target_name != f"{message_name}.proto":
                imports.add(target_name)
        elif field.type == 14:
            field_symbol = resolve_symbol(field.enum_type_symbol, symbol_to_name)
            target_name = resolve_file_name(
                field_symbol, region, variant_suffix_by_message_region
            )
            if target_name != f"{message_name}.proto":
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
    lines.append(f"message {message_name} {{")
    for field in fields:
        label = LABEL_MAP[field.label]
        field_type = resolve_type_name(
            field, region, variant_suffix_by_message_region, symbol_to_name
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
    symbol_to_name,
):
    output_dir = os.path.join(repo_root, "internal", "proto")
    os.makedirs(output_dir, exist_ok=True)
    outputs = []
    for symbol, message in en_data.messages.items():
        message_symbol = resolve_symbol(symbol, symbol_to_name)
        outputs.append((message_symbol, message, "EN", ""))
    for symbol, signature_groups in variant_groups.items():
        for signature, regions in signature_groups.items():
            region = regions[0]
            suffix = variant_suffix_by_message_region[symbol][region]
            message_symbol = resolve_symbol(symbol, symbol_to_name)
            message = region_data_map[region].messages.get(symbol)
            outputs.append((message_symbol, message, region, suffix))
    outputs.sort(key=lambda item: (item[2], item[0], item[3]))
    count = 0
    for symbol, message, region, suffix in outputs:
        file_name = f"{symbol}_{suffix}.proto" if suffix else f"{symbol}.proto"
        target_path = os.path.join(output_dir, file_name)
        log_progress(f"writing {os.path.relpath(target_path, repo_root)}")
        content = render_proto(
            symbol,
            message,
            region,
            variant_suffix_by_message_region,
            region_data_map[region].fields,
            suffix,
            symbol_to_name,
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
    symbol_to_name = {}
    for symbol, message in en_data.messages.items():
        symbol_to_name[symbol] = message.name.upper() if message.name else symbol
    variant_suffix_by_message_region, variant_groups = group_variants(
        en_signatures, other_signatures, symbol_to_name
    )
    generate_outputs(
        repo_root,
        en_data,
        region_data_map,
        variant_suffix_by_message_region,
        variant_groups,
        symbol_to_name,
    )


if __name__ == "__main__":
    main()
