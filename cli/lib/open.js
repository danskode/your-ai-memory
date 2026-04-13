import chalk from 'chalk';
import { spawn } from 'child_process';
import { findWiki, updateLastAccessed, readRegistry } from './registry.js';
import inquirer from 'inquirer';

export async function runOpen(nameArg) {
  let wiki;

  if (nameArg) {
    wiki = await findWiki(nameArg);
    if (!wiki) {
      console.log(chalk.red(`\n  No wiki found matching "${nameArg}".\n`));
      const registry = await readRegistry();
      if (registry.wikis.length > 0) {
        console.log(chalk.dim('  Registered wikis:'));
        for (const w of registry.wikis) {
          console.log(chalk.dim(`    ${w.name} — ${w.path}`));
        }
      }
      process.exit(1);
    }
  } else {
    const registry = await readRegistry();
    if (registry.wikis.length === 0) {
      console.log(chalk.yellow('\n  No wikis registered. Run `your-ai-memory create` first.\n'));
      process.exit(0);
    }

    const { selected } = await inquirer.prompt([
      {
        type: 'list',
        name: 'selected',
        message: 'Which wiki would you like to open?',
        choices: registry.wikis.map(w => ({
          name: `${w.name}  ${chalk.dim(w.domain)}`,
          value: w.name,
        })),
      },
    ]);
    wiki = await findWiki(selected);
  }

  await updateLastAccessed(wiki.name);

  console.log(chalk.dim(`\n  Opening ${chalk.white(wiki.name)} in Claude Code...\n`));
  const child = spawn('claude', [], {
    cwd: wiki.path,
    stdio: 'inherit',
    detached: false,
  });

  child.on('error', err => {
    if (err.code === 'ENOENT') {
      console.error(chalk.red('\n  Error: `claude` not found in PATH.'));
      console.error(chalk.dim('  Install Claude Code: https://claude.ai/code\n'));
    } else {
      console.error(chalk.red(`\n  Error launching Claude Code: ${err.message}\n`));
    }
    process.exit(1);
  });
}
