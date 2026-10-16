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
  // Feature modules mirror the backend hexagon: src/modules/<module>/<layer>/… where layer is one of
  // domain | application | ports | adapters | ui. These are null for the legacy flat layout.
  const fileModule = relative.match(/^modules\/([^/]+)\//)?.[1] ?? null;
  const fileLayer = relative.match(/^modules\/[^/]+\/([^/]+)\//)?.[1] ?? null;
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
      const resolvedRel = resolved ? path.relative(root, resolved) : null;
      const importedModule = resolvedRel?.match(/^modules\/([^/]+)\//)?.[1] ?? null;

      // Domain stays pure. Legacy flat domain imports nothing; a module's domain may reach only the
      // shared contracts zone (domain-safe cross-cutting types), nothing else.
      if (relative.startsWith('domain/')) failures.push(`${relative}: domain import ${specifier}`);
      if (fileLayer === 'domain' && !(resolvedRel && resolvedRel.startsWith('contracts/')))
        failures.push(`${relative}: module domain import ${specifier}`);

      // Application and ports point inward only. Legacy flat layout:
      if (relative.startsWith('application/') || relative.startsWith('ports/')) {
        const allowed = resolvedRel && /^(domain|ports|application)\//.test(resolvedRel);
        if (!allowed) failures.push(`${relative}: inward dependency violated by ${specifier}`);
      }
      // Module layout: the same module's domain/ports/application, or contracts; packages are fine.
      if (fileLayer === 'application' || fileLayer === 'ports') {
        const inward =
          resolvedRel === null ||
          /^contracts\//.test(resolvedRel) ||
          new RegExp(`^modules/${fileModule}/(domain|ports|application)/`).test(resolvedRel);
        if (!inward) failures.push(`${relative}: inward dependency violated by ${specifier}`);
      }

      // Contracts are the neutral zone: domain-safe, importing nothing but other contracts.
      if (relative.startsWith('contracts/') && resolvedRel && !resolvedRel.startsWith('contracts/'))
        failures.push(`${relative}: contract import ${specifier}`);

      // A module is reached from outside only through its index barrel — no deep cross-module import.
      // Dynamic import() for code-splitting is a CallExpression, not inspected here, so a lazy route
      // may still target a specific file inside a module.
      if (
        importedModule &&
        importedModule !== fileModule &&
        resolvedRel !== `modules/${importedModule}/index.ts`
      )
        failures.push(
          `${relative}: deep cross-module import ${specifier} (use modules/${importedModule})`,
        );

      if (
        resolved &&
        files.includes(resolved) &&
        !(ts.isImportDeclaration(node) && node.importClause?.isTypeOnly)
      )
        dependencies.push(resolved);
    }
    if (
      (/^(routes|presentation)\//.test(relative) || fileLayer === 'ui') &&
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
