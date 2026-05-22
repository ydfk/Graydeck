import { spawnSync } from "node:child_process";

function getCliOption(name) {
  const prefix = `--${name}=`;
  for (let index = 2; index < process.argv.length; index += 1) {
    const token = process.argv[index];
    if (token.startsWith(prefix)) {
      return token.slice(prefix.length);
    }
    if (token === `--${name}`) {
      const next = process.argv[index + 1];
      if (next && !next.startsWith("--")) {
        return next;
      }
    }
  }
  return undefined;
}

function getPositionalArgs() {
  const values = [];
  for (let index = 2; index < process.argv.length; index += 1) {
    const token = process.argv[index];
    if (token.startsWith("--")) {
      if (!token.includes("=") && process.argv[index + 1] && !process.argv[index + 1].startsWith("--")) {
        index += 1;
      }
      continue;
    }
    values.push(token);
  }
  return values;
}

function run(command, args) {
  const result = spawnSync(command, args, { stdio: "inherit", shell: true });
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}

const positional = getPositionalArgs();
const imageRepo = getCliOption("DOCKERHUB_REPO") ?? positional[0];
if (!imageRepo) {
  console.error("Missing image repo. Pass --DOCKERHUB_REPO=<repo> or first positional arg, e.g. ydfk/graydeck");
  process.exit(1);
}

const versionTag = getCliOption("DOCKER_IMAGE_TAG") ?? positional[1];
if (!versionTag || !isSemanticVersion(versionTag)) {
  console.error("Missing semantic version. Pass --DOCKER_IMAGE_TAG=1.0.1 or second positional arg.");
  process.exit(1);
}
const imageRef = `${imageRepo}:${versionTag}`;

console.log(`Building image: ${imageRef}`);
run("docker", ["build", "--build-arg", `VERSION=${versionTag}`, "-t", imageRef, "."]);
console.log(`Build completed: ${imageRef}`);

function isSemanticVersion(value) {
  return /^\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$/.test(value);
}
