import os
import tempfile
import subprocess
import importlib.util
import sys
import shutil


class ProtoCompiler:
    def __init__(self):
        self.temp_dirs = []  # Keep track of temp directories

    def compile_proto_dynamically(self, proto_content):
        """Dynamically compile proto file with improved temp file handling"""
        temp_dir = None
        try:
            # Create temporary directory for all files
            temp_dir = tempfile.mkdtemp(prefix="proto_compile_")
            self.temp_dirs.append(temp_dir)  # Keep reference to prevent early cleanup

            # Write proto content to file
            temp_proto_file = os.path.join(temp_dir, "dynamic_proto.proto")
            with open(temp_proto_file, 'w', encoding='utf-8') as f:
                f.write(proto_content)

            # Compile using protoc
            result = subprocess.run([
                'protoc',
                f'--python_out={temp_dir}',
                f'--proto_path={temp_dir}',
                'dynamic_proto.proto'
            ], capture_output=True, text=True, cwd=temp_dir)

            if result.returncode != 0:
                raise Exception(f"protoc failed: {result.stderr}")

            # Path to generated Python file
            temp_python_file = os.path.join(temp_dir, "dynamic_proto_pb2.py")

            if not os.path.exists(temp_python_file):
                # Try alternative naming (some protoc versions use different naming)
                possible_files = [f for f in os.listdir(temp_dir) if f.endswith('_pb2.py')]
                if not possible_files:
                    raise Exception(f"Generated Python file not found in {temp_dir}")
                temp_python_file = os.path.join(temp_dir, possible_files[0])

            # Add temp directory to Python path so imports work
            sys.path.insert(0, temp_dir)

            # Dynamically import compiled module
            module_name = "dynamic_proto_pb2"
            spec = importlib.util.spec_from_file_location(module_name, temp_python_file)
            dynamic_module = importlib.util.module_from_spec(spec)

            # Register the module to sys.modules so it can be imported elsewhere
            sys.modules[module_name] = dynamic_module
            spec.loader.exec_module(dynamic_module)

            # Find message classes (look for classes with DESCRIPTOR attribute)
            message_classes = {}
            for attr_name in dir(dynamic_module):
                attr = getattr(dynamic_module, attr_name)
                if hasattr(attr, 'DESCRIPTOR') and hasattr(attr, 'SerializeToString'):
                    message_classes[attr_name] = attr

            if not message_classes:
                raise Exception("No message classes found in proto file")

            # Use the first message class as default
            first_class_name = list(message_classes.keys())[0]
            message_descriptor = message_classes[first_class_name].DESCRIPTOR

            return {
                "module": dynamic_module,
                "message_classes": message_classes,
                "message_descriptor": message_descriptor
            }

        except Exception as e:
            # Cleanup on error
            if temp_dir and os.path.exists(temp_dir):
                shutil.rmtree(temp_dir, ignore_errors=True)
                if temp_dir in self.temp_dirs:
                    self.temp_dirs.remove(temp_dir)
            raise Exception(f"Proto compilation error: {str(e)}")

    def cleanup_temp_dirs(self):
        """Clean up all temporary directories"""
        for temp_dir in self.temp_dirs:
            if os.path.exists(temp_dir):
                shutil.rmtree(temp_dir, ignore_errors=True)
        self.temp_dirs.clear()