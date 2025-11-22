import os
import tempfile
import subprocess
import importlib.util

class ProtoCompiler:
    def __init__(self):
        pass

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
            spec = importlib.util.spec_from_file_location("dynamic_proto", temp_python_file)
            dynamic_module = importlib.util.module_from_spec(spec)
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

            # Cleanup temporary files
            os.unlink(temp_proto_file)
            # Keep the Python file for runtime

            return {
                "module": dynamic_module,
                "message_classes": message_classes,
                "message_descriptor": message_descriptor
            }

        except Exception as e:
            raise Exception(f"Proto compilation error: {str(e)}")