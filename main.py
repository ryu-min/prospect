import tkinter as tk
import customtkinter as ctk
from google.protobuf import message_factory
from google.protobuf import text_format
import os
import tempfile
import subprocess
from tkinter import ttk
from google.protobuf.internal import decoder
import struct

ctk.set_appearance_mode("Dark")
ctk.set_default_color_theme("blue")


class ProtoTreeView(ctk.CTkFrame):
    def __init__(self, master, **kwargs):
        super().__init__(master, **kwargs)

        # Configure style for treeview
        self.style = ttk.Style()
        self.style.configure("Treeview", background="#2b2b2b", foreground="white", fieldbackground="#2b2b2b")
        self.style.configure("Treeview.Heading", background="#3b3b3b", foreground="white")
        self.style.map("Treeview", background=[('selected', '#1f6aa5')])

        self.tree = ttk.Treeview(self, columns=("value", "type"), show="tree headings", height=20)
        self.tree.heading("#0", text="Field")
        self.tree.heading("value", text="Value")
        self.tree.heading("type", text="Type")

        # Configure column widths
        self.tree.column("#0", width=200, minwidth=150)
        self.tree.column("value", width=150, minwidth=100)
        self.tree.column("type", width=100, minwidth=80)

        # Scrollbars
        v_scroll = ttk.Scrollbar(self, orient="vertical", command=self.tree.yview)
        h_scroll = ttk.Scrollbar(self, orient="horizontal", command=self.tree.xview)
        self.tree.configure(yscrollcommand=v_scroll.set, xscrollcommand=h_scroll.set)

        # Grid layout
        self.tree.grid(row=0, column=0, sticky="nsew")
        v_scroll.grid(row=0, column=1, sticky="ns")
        h_scroll.grid(row=1, column=0, sticky="ew")

        self.grid_rowconfigure(0, weight=1)
        self.grid_columnconfigure(0, weight=1)

        # Bind double-click for editing
        self.tree.bind("<Double-1>", self.on_double_click)

        self.edit_entry = None

    def on_double_click(self, event):
        item = self.tree.selection()
        if not item:
            return

        item = item[0]
        column = self.tree.identify_column(event.x)

        if column == "#1":  # Value column
            # Cancel any existing edit
            if self.edit_entry:
                self.edit_entry.destroy()

            x, y, width, height = self.tree.bbox(item, column)
            values = self.tree.item(item, "values")
            if not values:
                return

            value = values[0]

            self.edit_entry = ctk.CTkEntry(self, width=width)
            self.edit_entry.place(x=x, y=y, width=width, height=height)
            self.edit_entry.insert(0, str(value))
            self.edit_entry.select_range(0, tk.END)
            self.edit_entry.focus()

            def save_edit(event=None):
                new_value = self.edit_entry.get()
                current_values = list(self.tree.item(item, "values"))
                current_values[0] = new_value
                self.tree.item(item, values=tuple(current_values))
                self.edit_entry.destroy()
                self.edit_entry = None

            def cancel_edit(event=None):
                if self.edit_entry:
                    self.edit_entry.destroy()
                    self.edit_entry = None

            self.edit_entry.bind("<Return>", save_edit)
            self.edit_entry.bind("<FocusOut>", cancel_edit)
            self.edit_entry.bind("<Escape>", cancel_edit)

    def clear_tree(self):
        """Clear all items from tree"""
        for item in self.tree.get_children():
            self.tree.delete(item)


class ProtoEditor(ctk.CTk):
    def __init__(self):
        super().__init__()

        self.title("prospect")
        self.geometry("900x600")

        # State variables
        self.current_proto_file = None
        self.message_descriptor = None
        self.message_instance = None
        self.binary_data = None
        self.parsed_data = None
        self.dynamic_module = None

        self.setup_ui()

    def setup_ui(self):
        # Main layout
        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(0, weight=1)

        # Left panel - navigation
        self.left_frame = ctk.CTkFrame(self, width=200)
        self.left_frame.grid(row=0, column=0, padx=10, pady=10, sticky="nsew")
        self.left_frame.grid_propagate(False)

        # Control buttons
        self.load_binary_btn = ctk.CTkButton(
            self.left_frame, text="Load Binary File", command=self.load_binary_file
        )
        self.load_binary_btn.pack(pady=5, fill="x")

        self.load_proto_btn = ctk.CTkButton(
            self.left_frame, text="Load .proto File", command=self.load_proto_file
        )
        self.load_proto_btn.pack(pady=5, fill="x")

        self.save_binary_btn = ctk.CTkButton(
            self.left_frame, text="Save Binary File", command=self.save_binary_file
        )
        self.save_binary_btn.pack(pady=5, fill="x")


        # Right panel - content
        self.right_frame = ctk.CTkFrame(self)
        self.right_frame.grid(row=0, column=1, padx=10, pady=10, sticky="nsew")
        self.right_frame.grid_columnconfigure(0, weight=1)
        self.right_frame.grid_rowconfigure(1, weight=1)

        # Tree view for structured display
        self.tree_view = ProtoTreeView(self.right_frame)
        self.tree_view.grid(row=0, column=0, sticky="nsew", padx=10, pady=10)

        # Status bar
        self.status_var = tk.StringVar(value="Ready")
        self.status_bar = ctk.CTkLabel(self, textvariable=self.status_var,
                                       font=ctk.CTkFont(size=12))
        self.status_bar.grid(row=1, column=0, columnspan=2, sticky="ew", padx=10, pady=5)

    def load_binary_file(self):
        from tkinter import filedialog
        file_path = filedialog.askopenfilename(
            title="Select binary proto file",
            filetypes=[("Binary files", "*.bin"), ("All files", "*.*")]
        )

        if file_path:
            try:
                with open(file_path, 'rb') as f:
                    self.binary_data = f.read()

                self.status_var.set(f"Loaded binary file: {os.path.basename(file_path)}")
                self.decode_and_display()

            except Exception as e:
                self.show_error(f"Error loading file: {str(e)}")

    def load_proto_file(self):
        from tkinter import filedialog
        file_path = filedialog.askopenfilename(
            title="Select .proto file",
            filetypes=[("Proto files", "*.proto"), ("All files", "*.*")]
        )

        if file_path:
            try:
                self.current_proto_file = file_path
                with open(file_path, 'r', encoding='utf-8') as f:
                    proto_content = f.read()

                # Dynamic proto compilation
                self.compile_proto_dynamically(proto_content)
                self.status_var.set(f"Loaded proto file: {os.path.basename(file_path)}")

                # Re-decode if we have binary data
                if self.binary_data:
                    self.decode_and_display()

            except Exception as e:
                self.show_error(f"Error loading proto file: {str(e)}")

    def compile_proto_dynamically(self, proto_content):
        """Dynamically compile proto file"""
        try:
            # Create temporary file
            with tempfile.NamedTemporaryFile(mode='w', suffix='.proto', delete=False) as f:
                f.write(proto_content)
                temp_proto_file = f.name

            # Create temp directory for output
            temp_dir = tempfile.mkdtemp()
            temp_python_file = os.path.join(temp_dir, "dynamic_proto_pb2.py")

            # Compile using protoc
            result = subprocess.run([
                'protoc',
                f'--python_out={temp_dir}',
                f'--proto_path={os.path.dirname(temp_proto_file)}',
                temp_proto_file
            ], capture_output=True, text=True)

            if result.returncode != 0:
                raise Exception(f"protoc failed: {result.stderr}")

            # Dynamically import compiled module
            import importlib.util
            spec = importlib.util.spec_from_file_location("dynamic_proto", temp_python_file)
            self.dynamic_module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(self.dynamic_module)

            # Find message classes (look for classes with DESCRIPTOR attribute)
            self.message_classes = {}
            for attr_name in dir(self.dynamic_module):
                attr = getattr(self.dynamic_module, attr_name)
                if hasattr(attr, 'DESCRIPTOR') and hasattr(attr, 'SerializeToString'):
                    self.message_classes[attr_name] = attr

            if not self.message_classes:
                raise Exception("No message classes found in proto file")

            # Use the first message class as default
            first_class_name = list(self.message_classes.keys())[0]
            self.message_descriptor = self.message_classes[first_class_name].DESCRIPTOR

            # Cleanup temporary files
            os.unlink(temp_proto_file)
            # Keep the Python file for runtime

        except Exception as e:
            raise Exception(f"Proto compilation error: {str(e)}")

    def decode_and_display(self):
        """Decode and display message"""
        try:
            if self.message_descriptor and self.binary_data:
                # Decode with proto schema
                message_class = message_factory.GetMessageClass(self.message_descriptor)
                self.message_instance = message_class()
                self.message_instance.ParseFromString(self.binary_data)

                # Formatted output
                display_text = text_format.MessageToString(self.message_instance)
                self.parsed_data = {"type": "proto", "data": self.message_instance, "text": display_text}

            elif self.binary_data:
                # Decode without schema (like --decode_raw)
                self.parsed_data = self.decode_raw()

            self.display_tree(self.parsed_data)

        except Exception as e:
            self.show_error(f"Decoding error: {str(e)}")

    def decode_raw(self):
        """Improved raw decoding that handles nested messages"""
        try:
            result = self.parse_message(self.binary_data, 0)
            formatted_text = self.format_raw_text(result["fields"])
            return {"type": "raw", "data": result["fields"], "text": formatted_text}
        except Exception as e:
            import traceback
            error_msg = f"Raw decoding error: {str(e)}\n{traceback.format_exc()}"
            self.show_error(error_msg)
            return {"type": "raw", "data": [], "text": f"Error: {str(e)}"}

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

    def display_tree(self, parsed_data):
        """Display data in tree view"""
        # Clear existing tree
        self.tree_view.clear_tree()

        if not parsed_data:
            return

        if parsed_data["type"] == "proto":
            self.display_proto_tree(parsed_data["data"])
        else:
            self.display_raw_tree(parsed_data["data"])

    def display_raw_tree(self, fields, parent=""):
        """Display raw decoded data in tree"""
        for field in sorted(fields, key=lambda x: x["field"]):
            field_id = field["field"]
            field_type = field["type"]
            value = field["value"]

            if field_type == "message":
                node = self.tree_view.tree.insert(parent, "end", text=f"Field {field_id}",
                                                  values=("", "message"))
                self.display_raw_tree(value, node)
            else:
                display_value = str(value)
                if field_type == "bytes" and "raw_data" in field:
                    display_value = f"bytes[{len(field['raw_data'])}]"

                self.tree_view.tree.insert(parent, "end", text=f"Field {field_id}",
                                           values=(display_value, field_type))

    def display_proto_tree(self, message, parent="", prefix=""):
        """Display proto message in tree"""
        if not message:
            return

        try:
            for field in message.DESCRIPTOR.fields:
                field_name = field.name
                if field.label == field.LABEL_REPEATED:
                    # Repeated field
                    repeated_values = getattr(message, field_name)
                    if repeated_values:
                        node = self.tree_view.tree.insert(parent, "end",
                                                          text=f"{prefix}{field_name}",
                                                          values=(f"{len(repeated_values)} items", "repeated"))
                        for i, item in enumerate(repeated_values):
                            if field.type == field.TYPE_MESSAGE:
                                self.display_proto_tree(item, node, f"[{i}].")
                            else:
                                self.tree_view.tree.insert(node, "end",
                                                           text=f"[{i}]",
                                                           values=(str(item), field.type_name))
                elif field.type == field.TYPE_MESSAGE:
                    # Nested message
                    nested_msg = getattr(message, field_name)
                    if nested_msg:
                        node = self.tree_view.tree.insert(parent, "end",
                                                          text=f"{prefix}{field_name}",
                                                          values=("", "message"))
                        self.display_proto_tree(nested_msg, node)
                    else:
                        self.tree_view.tree.insert(parent, "end",
                                                   text=f"{prefix}{field_name}",
                                                   values=("(empty)", "message"))
                else:
                    # Simple field
                    value = getattr(message, field_name)
                    self.tree_view.tree.insert(parent, "end",
                                               text=f"{prefix}{field_name}",
                                               values=(str(value), field.type_name))
        except Exception as e:
            self.tree_view.tree.insert(parent, "end",
                                       text="Error",
                                       values=(f"Display error: {str(e)}", "error"))

    def save_binary_file(self):
        """Save modified message as binary"""
        from tkinter import filedialog
        file_path = filedialog.asksaveasfilename(
            title="Save binary file",
            filetypes=[("Binary files", "*.bin"), ("All files", "*.*")],
            defaultextension=".bin"
        )

        if file_path:
            try:
                if self.message_instance:
                    output_data = self.message_instance.SerializeToString()
                else:
                    output_data = self.binary_data

                with open(file_path, 'wb') as f:
                    f.write(output_data)

                self.status_var.set(f"File saved: {os.path.basename(file_path)}")

            except Exception as e:
                self.show_error(f"Save error: {str(e)}")

    def show_error(self, message):
        self.status_var.set(f"Error: {message}")
        from tkinter import messagebox
        messagebox.showerror("Error", message)


if __name__ == "__main__":
    app = ProtoEditor()
    app.mainloop()