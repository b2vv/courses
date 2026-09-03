import fs from 'node:fs';
import path from 'node:path';
import ts from 'typescript';

const root = path.resolve('src');
function walk(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const file = path.join(dir, entry.name);
    return entry.isDirectory() ? walk(file) : /\.(ts|tsx)$/.test(file) ? [file] : [];
  });
}
const files = walk(root).filter(
  (file) =>
    !file.endsWith('.test.ts') &&
    !file.endsWith('.test.tsx') &&
    !file.includes('/generated/') &&
    !file.endsWith('routeTree.gen.ts'),
);
const edges = new Map();
const failures = [];
for (const file of files) {
  const relative = path.relative(root, file);
  const source = ts.createSourceFile(
    file,
    fs.readFileSync(file, 'utf8'),
    ts.ScriptTarget.Latest,
    true,
  );
  const dependencies = [];
  function visit(node) {
    if (
      (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) &&
      node.moduleSpecifier &&
      ts.isStringLiteral(node.moduleSpecifier)
    ) {
      const specifier = node.moduleSpecifier.text;
      const base = path.resolve(path.dirname(file), specifier);
      const resolved = [
        base + '.ts',
        base + '.tsx',
        path.join(base, 'index.ts'),
        path.join(base, 'index.tsx'),
      ].find((candidate) => fs.existsSync(candidate));
      if (relative.startsWith('domain/')) failures.push(`${relative}: domain import ${specifier}`);
      if (relative.startsWith('application/') || relative.startsWith('ports/')) {
        const allowed =
          resolved && /^(domain|ports|application)\//.test(path.relative(root, resolved));
        if (!allowed) failures.push(`${relative}: inward dependency violated by ${specifier}`);
      }
      if (
        resolved &&
        files.includes(resolved) &&
        !(ts.isImportDeclaration(node) && node.importClause?.isTypeOnly)
      )
        dependencies.push(resolved);
    }
    if (
      /^(routes|presentation)\//.test(relative) &&
      ts.isCallExpression(node) &&
      ts.isIdentifier(node.expression) &&
      node.expression.text === 'fetch'
    )
      failures.push(`${relative}: presentation calls fetch`);
    ts.forEachChild(node, visit);
  }
  visit(source);
  edges.set(file, dependencies);
}
const visited = new Set();
function inspect(file, stack) {
  if (stack.includes(file)) {
    failures.push(
      `runtime cycle: ${[...stack, file].map((f) => path.relative(root, f)).join(' -> ')}`,
    );
    return;
  }
  if (visited.has(file)) return;
  for (const dependency of edges.get(file) ?? []) inspect(dependency, [...stack, file]);
  visited.add(file);
}
for (const file of files) inspect(file, []);
if (failures.length) {
  process.stderr.write(failures.join('\n') + '\n');
  process.exitCode = 1;
} else process.stdout.write(`Architecture boundaries checked (${files.length} files).\n`);
