#!/usr/bin/env node
import chalk from 'chalk';
import inquirer from 'inquirer';
import { runCreate } from '../lib/create.js';
import { runList } from '../lib/list.js';
import { runOpen } from '../lib/open.js';

const [, , command, ...args] = process.argv;

async function interactiveMenu() {
  console.log(chalk.bold.cyan('\n  your-ai-memory\n'));
  const { action } = await inquirer.prompt([
    {
      type: 'list',
      name: 'action',
      message: 'What would you like to do?',
      choices: [
        { name: 'Create a new wiki', value: 'create' },
        { name: 'List all wikis', value: 'list' },
        { name: 'Open a wiki in Claude Code', value: 'open' },
        { name: 'Exit', value: 'exit' },
      ],
    },
  ]);
  return action;
}

async function main() {
  try {
    let cmd = command;

    if (!cmd) {
      cmd = await interactiveMenu();
    }

    switch (cmd) {
      case 'create':
        await runCreate();
        break;

      case 'list':
        await runList();
        break;

      case 'open':
        await runOpen(args[0]);
        break;

      case 'hub':
        // Delegate to Go TUI if available
        try {
          const { spawn } = await import('child_process');
          spawn('your-ai-memory-hub', [], { stdio: 'inherit' }).on('error', () => {
            console.error(chalk.red('\n  TUI hub not found.'));
            console.error(chalk.dim('  Install: go install github.com/nicolaieilstrup/your-ai-memory/tui/cmd/hub@latest\n'));
            process.exit(1);
          });
        } catch {
          console.error(chalk.red('\n  Could not launch hub.\n'));
        }
        break;

      case 'exit':
        break;

      default:
        console.error(chalk.red(`\n  Unknown command: ${cmd}`));
        console.log(chalk.dim('\n  Usage:'));
        console.log(chalk.dim('    your-ai-memory              # interactive menu'));
        console.log(chalk.dim('    your-ai-memory create       # scaffold a new wiki'));
        console.log(chalk.dim('    your-ai-memory list         # list all wikis'));
        console.log(chalk.dim('    your-ai-memory open [name]  # open in Claude Code'));
        console.log(chalk.dim('    your-ai-memory hub          # launch TUI hub\n'));
        process.exit(1);
    }
  } catch (err) {
    if (err.name === 'ExitPromptError') {
      // User cancelled with Ctrl+C
      console.log(chalk.dim('\n  Cancelled.\n'));
      process.exit(0);
    }
    console.error(chalk.red(`\n  Error: ${err.message}\n`));
    process.exit(1);
  }
}

main();
