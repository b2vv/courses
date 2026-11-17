// Fake Docker for the reset subprocess test; never calls a real Docker daemon.
import { readFileSync, writeFileSync } from 'node:fs';
const path = process.env.RESET_TEST_STATE;
const state = JSON.parse(readFileSync(path, 'utf8'));
const args = process.argv.slice(2);
state.calls.push(args);
function finish(code = 0) {
  writeFileSync(path, JSON.stringify(state));
  process.exit(code);
}
if (args[0] === 'compose') {
  if (
    JSON.stringify(args.slice(1, 5)) !==
    JSON.stringify(['-f', 'docker-compose.yml', '-f', 'docker-compose.e2e.yml'])
  )
    finish(90);
  if (JSON.stringify(args.slice(5)) === JSON.stringify(['down', '--remove-orphans'])) finish();
  if (JSON.stringify(args.slice(5)) !== JSON.stringify(['config', '--format', 'json'])) finish(91);
  console.log(JSON.stringify(state.config));
  finish();
}
if (args[0] === 'volume' && args[1] === 'ls' && args[2] === '--format' && args.length === 4) {
  state.lists++;
  if (state.failFirstList && state.lists === 1) finish(8);
  for (const [name, project, volume] of state.volumes) {
    console.log(
      args[3]
        .replaceAll('{{.Name}}', name)
        .replaceAll('{{.Label "com.docker.compose.project"}}', project)
        .replaceAll('{{.Label "com.docker.compose.volume"}}', volume),
    );
  }
  finish();
}
if (args[0] === 'volume' && args[1] === 'rm' && args.length > 2) {
  state.volumes = state.volumes.filter(([name]) => !args.slice(2).includes(name));
  finish();
}
console.error('Unexpected Docker command', args);
finish(92);
