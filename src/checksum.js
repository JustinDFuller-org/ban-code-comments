import crypto from "node:crypto";
import * as fs from "node:fs/promises";

function expectedChecksum(checksums, filename) {
  for (const line of String(checksums).split(/\r?\n/)) {
    const match = line.match(/^([0-9a-fA-F]{64})\s+\*?(.+)$/);
    if (match && match[2].trim() === filename) return match[1].toLowerCase();
  }
  throw new Error(`checksum entry not found for ${filename}`);
}

async function verifyChecksum(archivePath, checksums, filename) {
  const expected = expectedChecksum(checksums, filename);
  const contents = await fs.readFile(archivePath);
  const actual = crypto.createHash("sha256").update(contents).digest("hex");
  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${filename}: expected ${expected}, got ${actual}`);
  }
  return actual;
}

export { expectedChecksum, verifyChecksum };
