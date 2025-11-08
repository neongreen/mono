#!/usr/bin/env tsx

import { Project, Node } from 'ts-morph';
import * as path from 'path';

// Configuration for how to split the file
const SPLIT_CONFIG = {
  'types.ts': {
    types: ['AxisStatus', 'TkNote', 'RelationEdge', 'Relations', 'TkTask', 'TkGroup', 'TkProject', 'TkTreeItem', 'TkConfig'],
  },
  'utils.ts': {
    functions: ['escapeMarkdown', 'getNonce'],
  },
  'treeItems.ts': {
    classes: ['GroupTreeItem', 'TaskTreeItem'],
  },
  'detailProvider.ts': {
    classes: ['TaskDetailProvider'],
  },
  'decorationProvider.ts': {
    classes: ['TkDecorationProvider'],
  },
  'dragAndDropController.ts': {
    classes: ['TkDragAndDropController'],
  },
  'treeProvider.ts': {
    classes: ['TkProvider'],
  },
  'tkApi.ts': {
    functions: ['getTkConfig', 'fetchProjects', 'fetchTk'],
  },
  'commands.ts': {
    functions: ['rotateStatus', 'markDone', 'editTitle', 'createTask', 'quickCreateTask', 'createProject', 'deleteTask', 'deleteProject', 'moveTaskToGroup', 'updateToggleDoneButton'],
  },
};

async function main() {
  console.log('🔧 Starting refactoring of extension.ts...');

  // Create a new Project
  const project = new Project({
    tsConfigFilePath: path.join(__dirname, 'tsconfig.json'),
  });

  const extensionFile = project.getSourceFile('src/extension.ts');
  if (!extensionFile) {
    throw new Error('Could not find src/extension.ts');
  }

  console.log('📖 Loaded extension.ts');

  // Extract declarations
  const declarations = new Map<string, Node[]>();

  // Get all interfaces and type aliases
  const interfaces = extensionFile.getInterfaces();
  const typeAliases = extensionFile.getTypeAliases();
  const classes = extensionFile.getClasses();
  const functions = extensionFile.getFunctions();

  console.log(`Found: ${interfaces.length} interfaces, ${typeAliases.length} types, ${classes.length} classes, ${functions.length} functions`);

  // Group declarations by target file
  for (const [targetFile, config] of Object.entries(SPLIT_CONFIG)) {
    const nodes: Node[] = [];

    if (config.types) {
      for (const typeName of config.types) {
        const iface = interfaces.find(i => i.getName() === typeName);
        if (iface) nodes.push(iface);

        const typeAlias = typeAliases.find(t => t.getName() === typeName);
        if (typeAlias) nodes.push(typeAlias);
      }
    }

    if (config.classes) {
      for (const className of config.classes) {
        const cls = classes.find(c => c.getName() === className);
        if (cls) nodes.push(cls);
      }
    }

    if (config.functions) {
      for (const funcName of config.functions) {
        const func = functions.find(f => f.getName() === funcName);
        if (func) nodes.push(func);
      }
    }

    if (nodes.length > 0) {
      declarations.set(targetFile, nodes);
    }
  }

  console.log(`📦 Grouped declarations into ${declarations.size} files`);

  // Create new files
  for (const [targetFileName, nodes] of declarations.entries()) {
    const targetPath = path.join('src', targetFileName);
    console.log(`\n📝 Creating ${targetPath}...`);

    // Create new file
    const newFile = project.createSourceFile(targetPath, '', { overwrite: true });

    // Add imports from vscode
    const hasVscodeImport = nodes.some(node =>
      node.getText().includes('vscode.')
    );
    if (hasVscodeImport) {
      newFile.addImportDeclaration({
        moduleSpecifier: 'vscode',
        namespaceImport: 'vscode',
      });
    }

    // Add imports from node modules if needed
    const needsExecFile = nodes.some(node =>
      node.getText().includes('execFile')
    );
    if (needsExecFile) {
      newFile.addImportDeclaration({
        moduleSpecifier: 'node:child_process',
        namedImports: ['execFile'],
      });
      newFile.addImportDeclaration({
        moduleSpecifier: 'node:util',
        namedImports: ['promisify'],
      });
      newFile.addStatements('const execFileAsync = promisify(execFile);');
    }

    // Copy declarations to new file
    for (const node of nodes) {
      const text = node.getText();
      newFile.addStatements(text);

      // Make it exported
      const lastStatement = newFile.getStatements()[newFile.getStatements().length - 1];
      if (Node.isInterfaceDeclaration(lastStatement)) {
        lastStatement.setIsExported(true);
      } else if (Node.isTypeAliasDeclaration(lastStatement)) {
        lastStatement.setIsExported(true);
      } else if (Node.isClassDeclaration(lastStatement)) {
        lastStatement.setIsExported(true);
      } else if (Node.isFunctionDeclaration(lastStatement)) {
        lastStatement.setIsExported(true);
      }
    }

    console.log(`   ✓ Added ${nodes.length} declarations`);
  }

  // Update extension.ts: add imports and remove extracted code
  console.log('\n🔄 Updating extension.ts...');

  // Add imports for all new modules
  for (const [targetFileName] of declarations.entries()) {
    const moduleName = targetFileName.replace('.ts', '');
    const exportedNames = new Set<string>();

    const config = SPLIT_CONFIG[targetFileName as keyof typeof SPLIT_CONFIG];
    if (config.types) exportedNames.add(...config.types);
    if (config.classes) exportedNames.add(...config.classes);
    if (config.functions) exportedNames.add(...config.functions);

    extensionFile.addImportDeclaration({
      moduleSpecifier: `./${moduleName}`,
      namedImports: Array.from(exportedNames),
    });
  }

  // Remove extracted declarations from extension.ts
  for (const nodes of declarations.values()) {
    for (const node of nodes) {
      node.remove();
    }
  }

  console.log('   ✓ Added imports');
  console.log('   ✓ Removed extracted code');

  // Save all files
  console.log('\n💾 Saving files...');
  await project.save();

  console.log('✅ Refactoring complete!');
  console.log('\nCreated files:');
  for (const targetFileName of declarations.keys()) {
    console.log(`  - src/${targetFileName}`);
  }
  console.log('\n🔨 Next: Fix cross-module imports and run `npm run compile`');
}

main().catch(error => {
  console.error('❌ Error during refactoring:', error);
  process.exit(1);
});
