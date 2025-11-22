import customtkinter as ctk
from ui.main_window import ProtoEditor

ctk.set_appearance_mode("Dark")
ctk.set_default_color_theme("blue")

if __name__ == "__main__":
    app = ProtoEditor()
    app.mainloop()