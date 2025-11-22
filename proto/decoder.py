from google.protobuf import message_factory
from google.protobuf import text_format
from google.protobuf.internal import decoder
import struct

class ProtoDecoder:
    def __init__(self):
        pass

    def decode_with_schema(self, binary_data, message_descriptor):
        """Decode binary data using proto schema"""
        message_class = message_factory.GetMessageClass(message_descriptor)
        message_instance = message_class()
        message_instance.ParseFromString(binary_data)

        # Formatted output
        display_text = text_format.MessageToString(message_instance)

        return {
            "message_instance": message_instance,
            "display_text": display_text
        }

    def decode_raw(self, binary_data):
        """Improved raw decoding that handles nested messages"""
        try:
            result = self.parse_message(binary_data, 0)
            formatted_text = self.format_raw_text(result["fields"])
            return {"type": "raw", "data": result["fields"], "text": formatted_text}
        except Exception as e:
            import traceback
            error_msg = f"Raw decoding error: {str(e)}\n{traceback.format_exc()}"
            raise Exception(error_msg)

    def parse_message(self, data, start_pos):
        """Parse protobuf message recursively"""
        pos = start_pos
        fields = []

        while pos < len(data):
            try:
                # Read field tag
                if pos >= len(data):
                    break

                tag, new_pos = decoder._DecodeVarint(data, pos)
                if new_pos <= pos:  # Invalid position
                    break
                pos = new_pos

                field_number = tag >> 3
                wire_type = tag & 0x7

                if wire_type == 0:  # VARINT
                    value, new_pos = decoder._DecodeVarint(data, pos)
                    pos = new_pos
                    fields.append({
                        "field": field_number,
                        "type": "varint",
                        "value": value,
                        "wire_type": wire_type
                    })

                elif wire_type == 1:  # FIXED64
                    if pos + 8 > len(data):
                        break
                    value = struct.unpack('<Q', data[pos:pos + 8])[0]
                    pos += 8
                    fields.append({
                        "field": field_number,
                        "type": "fixed64",
                        "value": value,
                        "wire_type": wire_type
                    })

                elif wire_type == 2:  # LENGTH_DELIMITED
                    size, new_pos = decoder._DecodeVarint(data, pos)
                    pos = new_pos

                    if pos + size > len(data):
                        break

                    value_data = data[pos:pos + size]
                    pos += size

                    # Try to decode as nested message
                    nested_result = self.parse_message(value_data, 0)
                    if nested_result["fields"]:
                        fields.append({
                            "field": field_number,
                            "type": "message",
                            "value": nested_result["fields"],
                            "wire_type": wire_type,
                            "size": size,
                            "raw_data": value_data
                        })
                    else:
                        # Try to decode as string
                        try:
                            decoded_str = value_data.decode('utf-8')
                            # Check if it looks like a string (mostly printable characters)
                            printable_count = sum(1 for c in decoded_str if 32 <= ord(c) <= 126 or c in '\n\r\t')
                            if printable_count / len(decoded_str) > 0.8:  # 80% printable
                                fields.append({
                                    "field": field_number,
                                    "type": "string",
                                    "value": decoded_str,
                                    "wire_type": wire_type,
                                    "size": size
                                })
                            else:
                                fields.append({
                                    "field": field_number,
                                    "type": "bytes",
                                    "value": f"bytes[{size}]",
                                    "wire_type": wire_type,
                                    "size": size,
                                    "raw_data": value_data
                                })
                        except:
                            fields.append({
                                "field": field_number,
                                "type": "bytes",
                                "value": f"bytes[{size}]",
                                "wire_type": wire_type,
                                "size": size,
                                "raw_data": value_data
                            })

                elif wire_type == 5:  # FIXED32
                    if pos + 4 > len(data):
                        break
                    value = struct.unpack('<I', data[pos:pos + 4])[0]
                    pos += 4
                    fields.append({
                        "field": field_number,
                        "type": "fixed32",
                        "value": value,
                        "wire_type": wire_type
                    })

                else:
                    # Skip unknown wire types
                    continue

            except (IndexError, struct.error, ValueError) as e:
                # Skip to next field if we can't parse this one
                break

        return {"fields": fields, "end_pos": pos}

    def format_raw_text(self, fields, indent=0):
        """Format raw data as text similar to protoc --decode_raw"""
        text = ""
        prefix = "  " * indent

        for field in sorted(fields, key=lambda x: x["field"]):
            field_num = field["field"]
            field_type = field["type"]
            value = field["value"]

            if field_type == "message":
                text += f"{prefix}{field_num} {{\n"
                text += self.format_raw_text(value, indent + 1)
                text += f"{prefix}}}\n"
            else:
                text += f"{prefix}{field_num}: {value}\n"

        return text