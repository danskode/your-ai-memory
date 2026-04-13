import fs from 'fs-extra';
import path from 'path';
import os from 'os';

const CONFIG_DIR = path.join(os.homedir(), '.your-ai-memory');
const CONFIG_PATH = path.join(CONFIG_DIR, 'config.json');

export async function readRegistry() {
  try {
    await fs.ensureDir(CONFIG_DIR);
    if (!await fs.pathExists(CONFIG_PATH)) {
      return { wikis: [] };
    }
    return await fs.readJson(CONFIG_PATH);
  } catch {
    return { wikis: [] };
  }
}

export async function writeRegistry(data) {
  await fs.ensureDir(CONFIG_DIR);
  await fs.writeJson(CONFIG_PATH, data, { spaces: 2 });
}

export async function registerWiki(entry) {
  const registry = await readRegistry();
  const existing = registry.wikis.findIndex(w => w.path === entry.path);
  if (existing >= 0) {
    registry.wikis[existing] = { ...registry.wikis[existing], ...entry };
  } else {
    registry.wikis.push(entry);
  }
  await writeRegistry(registry);
}

export async function updateLastAccessed(name) {
  const registry = await readRegistry();
  const wiki = registry.wikis.find(
    w => w.name === name || w.name.toLowerCase() === name.toLowerCase()
  );
  if (wiki) {
    wiki.lastAccessed = new Date().toISOString().slice(0, 10);
    await writeRegistry(registry);
  }
  return wiki;
}

export async function findWiki(nameOrPath) {
  const registry = await readRegistry();
  return registry.wikis.find(
    w =>
      w.name === nameOrPath ||
      w.name.toLowerCase() === nameOrPath.toLowerCase() ||
      w.path === nameOrPath
  );
}
