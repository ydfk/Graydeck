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

function buildTimestampTag(now = new Date()) {
  const pad = (value) => String(value).padStart(2, "0");
  return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}${pad(now.getHours())}${pad(now.getMinutes())}`;
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

const versionTag = getCliOption("DOCKER_IMAGE_TAG") ?? positional[1] ?? buildTimestampTag();
const imageRef = `${imageRepo}:${versionTag}`;

console.log(`Building image: ${imageRef}`);
run("docker", ["build", "-t", imageRef, "."]);
console.log(`Build completed: ${imageRef}`);
