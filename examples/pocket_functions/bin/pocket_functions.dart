import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;
import 'package:yaml/yaml.dart';
import 'package:yaml_writer/yaml_writer.dart';

void main(List<String> args) async {
  // var options = parseOptions(args);
  final file = File('pubspec.yaml');
  final yamlContent = file.readAsStringSync();
  final doc = loadYaml(yamlContent);

  final packageName = doc['name'];
  Map functionDeps = Map.of(doc['dependencies'] ?? {});
  final pocketFunctionConfig = doc['pocket_functions'] ?? {};

  print("Pocket Functions Deploy for $packageName\n");

  var functionFile = _getFunctionFile();
  if (functionFile != null) {
    functionDeps.remove("pocket_functions");
    var functionPath =
        pocketFunctionConfig["path"] ?? "/${packageName.replaceAll("_", "-")}";
    await _createFunction(
        functionFile, functionPath, _escapeCode(_toYaml(functionDeps)));
  } else {
    print("Stopping. No function defined in lib/");
  }
}

String _toYaml(Map map) {
  var yamlWriter = YamlWriter();

  return yamlWriter.write(map);
}

Future<void> _createFunction(
    File functionFile, String functionName, String functionDeps) async {
  print("Starting deploy to $functionName ...\n");

  var payload = _createPayload(
    path: functionName,
    code: _getFunctionContent(functionFile),
    deps: functionDeps,
  );

  final response = await http.post(
    Uri.parse("http://localhost:8080/_/create"),
    headers: {"Content-Type": "application/json"},
    body: jsonEncode(payload),
  );

  // Manejar la respuesta
  if (response.statusCode == 200) {
    print("Deploy Succeed");
  } else {
    print("Deploy Error: ${response.statusCode}, ${response.body}");
  }
}

Map<String, String> _createPayload({
  required String path,
  required String code,
  required String deps,
}) {
  return {
    "path": path,
    "code": _escapeCode(code),
    "deps": deps,
  };
}

String _escapeCode(String code) {
  final escapedCode = jsonEncode(code);
  return jsonDecode(escapedCode);
}

String _getFunctionContent(File file) {
  return file.readAsStringSync();
}

File? _getFunctionFile() {
  final directory = Directory('lib');
  if (!directory.existsSync()) {
    return null;
  }

  FileSystemEntity? dartFile = directory
      .listSync()
      .where(
        (file) => file is File && file.path.endsWith('.dart'),
      )
      .firstOrNull;

  if (dartFile != null && dartFile is File) {
    return dartFile;
  }

  return null;
}
