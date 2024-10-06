FROM dart:stable AS build

# Resolve app dependencies.
WORKDIR /app
COPY pubspec.* ./
RUN mkdir -p /app/vendor

ARG name

# Copy and extract library
# TODO: Function name should be a template
RUN mkdir -p /app/vendor/dart_function
COPY ./$name.zip /app/vendor/$name/
RUN cd /app/vendor/$name && \
	unzip $name.zip
RUN cd /app/vendor/$name && ls && dart run build_runner build


# Install all dependencies
RUN cd /app && awk '{sub(/%s/, ENVIRON["name"]); sub(/%s/, ENVIRON["name"]); print}' pubspec.yaml.template > pubspec.yaml
RUN dart pub get

# Copy app source code and AOT compile it.
COPY . .
# Ensure packages are still up-to-date if anything has changed
RUN dart pub get --offline
RUN cd /app/bin && awk '{sub(/%s/, ENVIRON["name"]); sub(/%s/, ENVIRON["name"]); print}' executor.dart.template > executor.dart
RUN dart compile exe bin/executor.dart -o bin/server

# Build minimal serving image from AOT-compiled `/server` and required system
# libraries and configuration files stored in `/runtime/` from the build stage.
FROM scratch
COPY --from=build /runtime/ /
COPY --from=build /app/bin/server /app/bin/

# Start server.
EXPOSE 8080
CMD ["/app/bin/server"]
