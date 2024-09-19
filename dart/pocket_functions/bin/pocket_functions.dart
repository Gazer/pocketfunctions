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

  var id = await _createFunction(functionPath);
  await _deployFunction(id, zipFileName);
}

Future<int> _createFunction(
  String functionPath,
) async {
  print("Starting deploy to $functionPath ...\n");

  final uri = Uri.parse("http://localhost:8080/_/create");
  var response = await http.post(
    uri,
    headers: <String, String>{
      'Content-Type': 'application/json; charset=UTF-8',
    },
    body: jsonEncode(<String, String>{
      'uri': functionPath,
    }),
  );

  if (response.statusCode == 200) {
    var id = jsonDecode(response.body)["id"];
    print("Function created with id=$id");
    return id;
  } else {
    print("Deploy Error: ${response.statusCode}, ${response.body}");
    exit(1);
  }
}

Future<void> _deployFunction(
  int id,
  String zipFilePath,
) async {
  final uri = Uri.parse("http://localhost:8080/_/upload/$id");
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
