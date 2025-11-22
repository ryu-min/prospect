import tkinter as tk
import customtkinter as ctk
from tkinter import ttk

class ProtoTreeView(ctk.CTkFrame):
    def __init__(self, master, **kwargs):
        super().__init__(master, **kwargs)
        self.setup_treeview()
        self.edit_entry = None

    def setup_treeview(self):
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