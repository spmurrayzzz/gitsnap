const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const gitsnap = require('../node');

async function main() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'gitsnap-node-'));
  const worktree = path.join(root, 'worktree');
  const home = path.join(root, 'home');

  fs.mkdirSync(worktree);
  fs.mkdirSync(home);
  process.env.GITSNAP_HOME = home;
  fs.writeFileSync(path.join(worktree, 'a.txt'), 'hello\n');

  await gitsnap.init({ worktree });
  const hash = await gitsnap.save({ worktree, alias: 'smoke' });

  if (!/^[0-9a-f]{40}$/.test(hash)) {
    throw new Error(`invalid snapshot hash: ${hash}`);
  }

  const resolved = await gitsnap.resolve('smoke', { worktree });
  if (resolved !== hash) {
    throw new Error(`resolve returned ${resolved}, expected ${hash}`);
  }

  fs.writeFileSync(path.join(worktree, 'a.txt'), 'hello again\n');

  const files = await gitsnap.files('smoke', { worktree });
  if (!files.includes('a.txt')) {
    throw new Error(`changed files missing a.txt: ${JSON.stringify(files)}`);
  }

  const patch = await gitsnap.diff('smoke', { worktree });
  if (!patch.includes('hello again')) {
    throw new Error('diff did not include updated content');
  }

  await gitsnap.restore('smoke', { worktree, paths: ['a.txt'] });
  const restored = fs.readFileSync(path.join(worktree, 'a.txt'), 'utf8');
  if (restored !== 'hello\n') {
    throw new Error(`restore wrote ${JSON.stringify(restored)}`);
  }

  const aliases = await gitsnap.aliases({ worktree });
  if (aliases.length !== 1 || aliases[0].name !== 'smoke') {
    throw new Error(`unexpected aliases: ${JSON.stringify(aliases)}`);
  }

  const syncHash = gitsnap.sync.resolve('smoke', { worktree });
  if (syncHash !== hash) {
    throw new Error(`sync resolve returned ${syncHash}, expected ${hash}`);
  }

  process.chdir(worktree);
  const pending = gitsnap.save({ worktree: '.', alias: 'queued' });
  try {
    gitsnap.sync.resolve('smoke', { worktree });
    throw new Error('sync call did not reject during pending async work');
  } catch (err) {
    if (!err.message.includes('async operation already pending')) {
      throw err;
    }
  }
  await pending;

  await gitsnap.cleanup({ worktree });
  fs.rmSync(root, { recursive: true, force: true });
  console.log('node smoke ok');
}

main().catch(err => {
  console.error(err);
  process.exitCode = 1;
});
