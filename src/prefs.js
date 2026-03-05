import fs from 'node:fs';
import path from 'node:path';

function prefsPath() {
  return path.join(process.cwd(), '.prefs.json');
}

export function loadPrefs() {
  if (!fs.existsSync(prefsPath())) return {};
  try {
    return JSON.parse(fs.readFileSync(prefsPath(), 'utf8'));
  } catch {
    return {};
  }
}

export function savePrefs(prefs) {
  const current = loadPrefs();
  const merged = { ...current, ...prefs };
  fs.writeFileSync(prefsPath(), JSON.stringify(merged, null, 2), 'utf8');
}
