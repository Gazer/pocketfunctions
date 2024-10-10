FROM dart:stable AS build

ARG name

# Resolve app dependencies.
WORKDIR /app
COPY ./dist/$name.zip .
RUN cd /app && unzip $name.zip

# Install all dependencies
RUN dart pub get

# Compile
RUN mkdir -p bin/
RUN dart compile exe lib/*.dart -o bin/server

# Build minimal serving image from AOT-compiled `/server` and required system
# libraries and configuration files stored in `/runtime/` from the build stage.
FROM scratch
COPY --from=build /runtime/ /
COPY --from=build /app/bin/server /app/bin/

# Start server.
EXPOSE 8080
CMD ["/app/bin/server"]
