import tkinter as tk
import customtkinter as ctk
import os
from tkinter import filedialog, messagebox

from ui.tree_view import ProtoTreeView
from proto.compiler import ProtoCompiler
from proto.decoder import ProtoDecoder


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

        # Initialize components
        self.proto_compiler = ProtoCompiler()
        self.proto_decoder = ProtoDecoder()

        self.setup_ui()

    def setup_ui(self):
        # Configure main grid layout
        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)  # Main content area
        self.grid_rowconfigure(2, weight=0)  # Status bar

        # Toolbar frame at the top
        self.toolbar_frame = ctk.CTkFrame(self, height=40)
        self.toolbar_frame.grid(row=0, column=0, sticky="ew", padx=10, pady=5)
        self.toolbar_frame.grid_propagate(False)

        # Left-aligned container for buttons
        self.buttons_container = ctk.CTkFrame(self.toolbar_frame, fg_color="transparent")
        self.buttons_container.pack(side="left", fill="y", padx=5, pady=5)

        # Toolbar buttons - packed to the left inside container
        self.load_binary_btn = ctk.CTkButton(
            self.buttons_container,
            text="open binary",
            command=self.load_binary_file,
            width=100
        )
        self.load_binary_btn.pack(side="left", padx=5)

        self.load_proto_btn = ctk.CTkButton(
            self.buttons_container,
            text="open schema",
            command=self.load_proto_file,
            width=100
        )
        self.load_proto_btn.pack(side="left", padx=5)

        self.save_binary_btn = ctk.CTkButton(
            self.buttons_container,
            text="save",
            command=self.save_binary_file,
            width=60
        )
        self.save_binary_btn.pack(side="left", padx=5)

        # Main content area
        self.content_frame = ctk.CTkFrame(self)
        self.content_frame.grid(row=1, column=0, padx=10, pady=5, sticky="nsew")
        self.content_frame.grid_columnconfigure(0, weight=1)
        self.content_frame.grid_rowconfigure(0, weight=1)

        # Tree view for structured display - now fills all available space
        self.tree_view = ProtoTreeView(self.content_frame)
        self.tree_view.grid(row=0, column=0, sticky="nsew")

        # Status bar at the bottom
        self.status_var = tk.StringVar(value="Ready")
        self.status_bar = ctk.CTkLabel(
            self,
            textvariable=self.status_var,
            font=ctk.CTkFont(size=12),
            height=20
        )
        self.status_bar.grid(row=2, column=0, sticky="ew", padx=10, pady=5)

    def load_binary_file(self):
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
                result = self.proto_compiler.compile_proto_dynamically(proto_content)
                self.dynamic_module = result["module"]
                self.message_classes = result["message_classes"]
                self.message_descriptor = result["message_descriptor"]

                self.status_var.set(f"Loaded proto file: {os.path.basename(file_path)}")

                # Re-decode if we have binary data
                if self.binary_data:
                    self.decode_and_display()

            except Exception as e:
                self.show_error(f"Error loading proto file: {str(e)}")

    def decode_and_display(self):
        """Decode and display message"""
        try:
            if self.message_descriptor and self.binary_data:
                # Decode with proto schema
                result = self.proto_decoder.decode_with_schema(
                    self.binary_data, self.message_descriptor
                )
                self.message_instance = result["message_instance"]
                self.parsed_data = {
                    "type": "proto",
                    "data": self.message_instance,
                    "text": result["display_text"]
                }

            elif self.binary_data:
                # Decode without schema
                self.parsed_data = self.proto_decoder.decode_raw(self.binary_data)

            self.display_tree(self.parsed_data)

        except Exception as e:
            self.show_error(f"Decoding error: {str(e)}")

    def display_tree(self, parsed_data):
        """Display data in tree view"""
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
        messagebox.showerror("Error", message)