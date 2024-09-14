import 'dart:convert';
import 'dart:io';
import 'package:path/path.dart' as p;

import 'package:http/http.dart' as http;
import 'package:pocket_functions/zipper.dart';
import 'package:yaml/yaml.dart';
import 'package:yaml_writer/yaml_writer.dart';

void main(List<String> args) async {
  // var options = parseOptions(args);
  final file = File('pubspec.yaml');
  final yamlContent = file.readAsStringSync();
  final doc = loadYaml(yamlContent);

  final packageName = doc['name'];
  final pocketFunctionConfig = doc['pocket_functions'] ?? {};

  print("Pocket Functions Deploy for $packageName\n");
  var functionPath =
      pocketFunctionConfig["path"] ?? "/${packageName.replaceAll("_", "-")}";

  var zipFileName = "$packageName.zip";
  await createPackageZip(".", zipFileName);

  await _createFunction(functionPath, zipFileName);
}

Future<void> _createFunction(
  String functionName,
  String zipFilePath,
) async {
  print("Starting deploy to $functionName ...\n");

  if (functionName[0] == "/") {
    functionName = functionName.substring(1);
  }

  final uri = Uri.parse("http://localhost:8080/_/create/$functionName");
  final request = http.MultipartRequest('POST', uri);

  final fileName = p.basename(zipFilePath);
  request.files.add(await http.MultipartFile.fromPath(
    'file',
    zipFilePath,
    filename: fileName,
  ));

  final response = await request.send();
  final responseBody = await response.stream.bytesToString();

  if (response.statusCode == 200) {
    print("Deploy Succeed");
  } else {
    print("Deploy Error: ${response.statusCode}, $responseBody");
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
