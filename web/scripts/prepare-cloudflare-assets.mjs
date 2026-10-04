import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";

const sitePath = process.argv[2];

if (!sitePath || !/^[a-z0-9-]+$/.test(sitePath)) {
  throw new Error("Expected a URL-safe site path argument.");
}

const source = path.resolve("out");
const destinationRoot = path.resolve(".cloudflare/assets");
const destination = path.join(destinationRoot, sitePath);

await rm(destinationRoot, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });

// Wrangler only reads _redirects from the assets root (a nested copy is served as
// a plain file), so move it there and prefix each rule's paths with the site path.
const nestedRedirects = path.join(destination, "_redirects");
const rules = await readFile(nestedRedirects, "utf8").catch((error) => {
  if (error.code === "ENOENT") return null;
  throw error;
});
if (rules !== null) {
  const prefixed = rules.split("\n").map((line) => {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) return line;
    return trimmed
      .split(/\s+/)
      .map((part, index) => (index < 2 && part.startsWith("/") ? `/${sitePath}${part}` : part))
      .join(" ");
  });
  await writeFile(path.join(destinationRoot, "_redirects"), prefixed.join("\n"));
  await rm(nestedRedirects);
}
