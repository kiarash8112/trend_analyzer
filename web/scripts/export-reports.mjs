// Publishes ../reports as static JSON in the build output, standing in for
// cmd/report_server on static hosting. public/_redirects maps /api/reports to
// these files. Run from web/ after `ng build`.
import { mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';

const out = 'dist/web/browser';
const pattern = /^(\d{4}-\d{2}-\d{2})-(daily|weekly|monthly)$/;
mkdirSync(`${out}/data`, { recursive: true });

const list = [];
for (const file of readdirSync('../reports')) {
  if (!file.endsWith('.json')) continue;
  const name = file.slice(0, -5);
  const match = name.match(pattern);
  if (!match) continue;
  const raw = readFileSync(`../reports/${file}`, 'utf8');
  writeFileSync(`${out}/data/${file}`, raw);
  list.push({
    name,
    date: match[1],
    since: match[2],
    repo_count: JSON.parse(raw).length,
  });
}
list.sort((a, b) => b.name.localeCompare(a.name));
writeFileSync(`${out}/data/index.json`, JSON.stringify(list));
console.log(`exported ${list.length} report(s) to ${out}/data`);
