import inquirer from 'inquirer';
import chalk from 'chalk';
import fs from 'fs-extra';
import path from 'path';
import os from 'os';
import Handlebars from 'handlebars';
import { fileURLToPath } from 'url';
import { spawn } from 'child_process';
import { registerWiki } from './registry.js';

// Register a simple join helper so templates can do {{join array ", "}}
Handlebars.registerHelper('join', (arr, sep) =>
  Array.isArray(arr) ? arr.join(typeof sep === 'string' ? sep : ', ') : String(arr ?? '')
);

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const TEMPLATE_DIR = path.join(__dirname, '..', '..', 'template');

const SOURCE_CHOICES = [
  { name: 'Articles (blog posts, web pages)', value: 'articles' },
  { name: 'Papers (research, academic)', value: 'papers' },
  { name: 'Videos / Transcripts', value: 'videos' },
  { name: 'Books', value: 'books' },
  { name: 'Personal notes', value: 'notes' },
  { name: 'Documentation / specs', value: 'docs' },
];

const GOAL_CHOICES = [
  { name: 'Learning a new domain', value: 'learning' },
  { name: 'Ongoing research', value: 'research' },
  { name: 'Professional reference', value: 'professional reference' },
  { name: 'Sharing expertise with others', value: 'expertise-sharing' },
];

export async function runCreate() {
  console.log(chalk.bold.cyan('\n  your-ai-memory — Create a new wiki\n'));

  const answers = await inquirer.prompt([
    {
      type: 'input',
      name: 'name',
      message: 'Wiki name (used as directory name):',
      validate: v => v.trim().length > 0 || 'Name is required',
    },
    {
      type: 'input',
      name: 'location',
      message: answers => `Location (default: ~/wikis/${answers.name}):`,
      default: answers => path.join(os.homedir(), 'wikis', answers.name),
    },
    {
      type: 'input',
      name: 'domain',
      message: 'Domain — what topic will this wiki cover?',
      validate: v => v.trim().length > 0 || 'Domain is required',
    },
    {
      type: 'list',
      name: 'goal',
      message: 'Primary goal:',
      choices: GOAL_CHOICES,
    },
    {
      type: 'checkbox',
      name: 'sources',
      message: 'Source types you will ingest (space to toggle):',
      choices: SOURCE_CHOICES,
      validate: v => v.length > 0 || 'Select at least one source type',
    },
    {
      type: 'input',
      name: 'topics',
      message: 'Topic tags (comma-separated, e.g. computer-science, personal):',
      default: 'general',
      filter: v => v.split(',').map(t => t.trim()).filter(Boolean),
    },
    {
      type: 'input',
      name: 'language',
      message: 'Language for wiki pages:',
      default: 'English',
    },
  ]);

  const wikiPath = answers.location.replace(/^~/, os.homedir());
  const today = new Date().toISOString().slice(0, 10);

  console.log(chalk.dim(`\n  Scaffolding ${chalk.white(wikiPath)} ...\n`));

  await scaffoldWiki({ ...answers, wikiPath, today });

  await registerWiki({
    name: answers.name,
    path: wikiPath,
    domain: answers.domain,
    topics: answers.topics,
    created: today,
    lastAccessed: today,
  });

  printOnboarding(answers, wikiPath);

  const { openNow } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'openNow',
      message: 'Open this wiki in Claude Code now?',
      default: true,
    },
  ]);

  if (openNow) {
    console.log(chalk.dim(`\n  Launching Claude Code in ${wikiPath} ...\n`));
    spawn('claude', [], { cwd: wikiPath, stdio: 'inherit', detached: false });
  }
}

async function scaffoldWiki({ name, domain, goal, sources, topics, language, wikiPath, today }) {
  // Read and compile CLAUDE.md template
  const tmplSrc = await fs.readFile(path.join(TEMPLATE_DIR, 'CLAUDE.md.tmpl'), 'utf8');
  const tmpl = Handlebars.compile(tmplSrc);
  const claudeMd = tmpl({ name, domain, goal, sources, topics, language, created: today });

  await fs.ensureDir(wikiPath);
  await fs.writeFile(path.join(wikiPath, 'CLAUDE.md'), claudeMd);

  // .claude/commands/ingest.md
  const ingestSrc = await fs.readFile(
    path.join(TEMPLATE_DIR, '.claude', 'commands', 'ingest.md'),
    'utf8'
  );
  await fs.ensureDir(path.join(wikiPath, '.claude', 'commands'));
  await fs.writeFile(path.join(wikiPath, '.claude', 'commands', 'ingest.md'), ingestSrc);

  // wiki/ stubs
  const wikiTmplDir = path.join(TEMPLATE_DIR, 'wiki');
  await fs.ensureDir(path.join(wikiPath, 'wiki'));
  for (const file of ['index.md', 'log.md', 'overview.md']) {
    const src = await fs.readFile(path.join(wikiTmplDir, file), 'utf8');
    const rendered = Handlebars.compile(src)({ name, domain, created: today });
    await fs.writeFile(path.join(wikiPath, 'wiki', file), rendered);
  }

  // wiki/ subdirectory placeholders
  for (const dir of ['concepts', 'patterns', 'papers', 'people', 'connections', 'questions']) {
    await fs.ensureDir(path.join(wikiPath, 'wiki', dir));
  }

  // raw/ subdirectories — only chosen source types
  for (const source of sources) {
    await fs.ensureDir(path.join(wikiPath, 'raw', source));
  }
}

function printOnboarding(answers, wikiPath) {
  console.log(chalk.green.bold('\n  Wiki created successfully!\n'));
  console.log(chalk.bold('  What now:\n'));
  console.log(`  ${chalk.cyan('1.')} Drop source material into ${chalk.white(`raw/`)}`);
  console.log(`     ${chalk.dim('Supported: ' + answers.sources.join(', '))}`);
  console.log(`\n  ${chalk.cyan('2.')} Open the wiki in Claude Code:`);
  console.log(`     ${chalk.white(`cd ${wikiPath} && claude`)}`);
  console.log(`\n  ${chalk.cyan('3.')} Ingest a source with the slash command:`);
  console.log(`     ${chalk.white('/ingest articles/my-article.md')}`);
  console.log(`\n  ${chalk.cyan('4.')} After a few ingests, update the overview:`);
  console.log(`     ${chalk.white('/update-overview')}`);
  console.log(`\n  ${chalk.dim('Tip: run')} ${chalk.white('your-ai-memory list')} ${chalk.dim('to see all your wikis.')}\n`);
}
