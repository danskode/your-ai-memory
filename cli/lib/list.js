import chalk from 'chalk';
import { readRegistry } from './registry.js';

export async function runList() {
  const registry = await readRegistry();

  if (registry.wikis.length === 0) {
    console.log(chalk.yellow('\n  No wikis registered yet. Run `your-ai-memory create` to get started.\n'));
    return;
  }

  // Group by first topic tag
  const groups = {};
  for (const wiki of registry.wikis) {
    const tag = (wiki.topics && wiki.topics[0]) || 'untagged';
    if (!groups[tag]) groups[tag] = [];
    groups[tag].push(wiki);
  }

  console.log(chalk.bold.cyan('\n  your-ai-memory — Registered Wikis\n'));

  for (const [tag, wikis] of Object.entries(groups)) {
    console.log(chalk.bold.yellow(`  [${tag}]`));

    for (const wiki of wikis) {
      const name = chalk.white(wiki.name.padEnd(20));
      const domain = chalk.dim(wiki.domain.padEnd(35));
      const accessed = chalk.dim(wiki.lastAccessed || wiki.created || '—');
      console.log(`    ${name} ${domain} ${accessed}`);
      console.log(`    ${chalk.dim(wiki.path)}`);
      console.log();
    }
  }

  console.log(chalk.dim(`  ${registry.wikis.length} wiki(s) total.\n`));
}
