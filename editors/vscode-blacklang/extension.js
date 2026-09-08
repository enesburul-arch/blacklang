"use strict";

const childProcess = require("child_process");
const path = require("path");

let vscode = null;
try {
  vscode = require("vscode");
} catch {
  vscode = null;
}

let diagnosticCollection = null;

function activate(context) {
  if (!vscode) {
    return;
  }

  diagnosticCollection = vscode.languages.createDiagnosticCollection("blacklang");
  context.subscriptions.push(diagnosticCollection);

  let manifestPromise = null;
  const loadManifest = (document) => {
    if (!manifestPromise) {
      manifestPromise = runBlackJSON(["ide", "--json"], workspaceFolderFor(document)).catch((error) => {
        manifestPromise = null;
        throw error;
      });
    }
    return manifestPromise;
  };

  context.subscriptions.push(
    vscode.languages.registerCompletionItemProvider(
      [{ language: "blacklang" }, { language: "blacklang-theme" }],
      {
        async provideCompletionItems(document) {
          const manifest = await loadManifest(document);
          return [
            ...completionItemsFromManifest(manifest),
            ...snippetItemsFromManifest(manifest),
          ];
        },
      },
      " ",
      "\n"
    )
  );

  context.subscriptions.push(
    vscode.languages.registerCodeActionsProvider(
      { language: "blacklang" },
      {
        provideCodeActions(document, range, context) {
          return codeActionsForDiagnostics(document, context.diagnostics);
        },
      },
      {
        providedCodeActionKinds: [
          vscode.CodeActionKind.QuickFix,
          vscode.CodeActionKind.SourceFixAll,
        ],
      }
    )
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("blacklang.refreshDiagnostics", async () => {
      await refreshDiagnosticsForOpenDocuments();
      vscode.window.showInformationMessage("BlackLang diagnostics refreshed.");
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("blacklang.showIDEManifest", async () => {
      const editor = vscode.window.activeTextEditor;
      const manifest = await loadManifest(editor?.document);
      await showJSONDocument("BlackLang IDE Manifest", manifest);
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("blacklang.inspectAffected", async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || !isBlackLangSourceDocument(editor.document)) {
        vscode.window.showWarningMessage("Open a .black source file before inspecting an affected symbol.");
        return;
      }
      const symbol = symbolAtSelection(editor);
      if (!symbol) {
        vscode.window.showWarningMessage("Place the cursor on a BlackLang symbol first.");
        return;
      }
      const payload = await runBlackJSON(["inspect", editor.document.fileName, "--affected", symbol, "--json"], workspaceFolderFor(editor.document));
      await showJSONDocument(`BlackLang Affected: ${symbol}`, payload);
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("blacklang.formatSource", async (uri) => {
      const target = uri?.fsPath ?? vscode.window.activeTextEditor?.document.fileName;
      if (!target) {
        vscode.window.showWarningMessage("Open a .black source file before formatting.");
        return;
      }
      await runBlackJSON(["format", target, "--json"], workspaceFolderForPath(target));
      await refreshDiagnosticsForPath(target);
      vscode.window.showInformationMessage("BlackLang source formatted.");
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("blacklang.showDiagnosticInfo", async (diagnostic) => {
      const code = diagnostic?.code ?? "BlackLang diagnostic";
      const suggestion = diagnostic?.suggestion ? `\n\nSuggestion: ${diagnostic.suggestion}` : "";
      vscode.window.showInformationMessage(`${code}: ${diagnostic?.message ?? "See BlackLang diagnostics."}${suggestion}`);
    })
  );

  context.subscriptions.push(
    vscode.workspace.onDidOpenTextDocument((document) => refreshDiagnosticsForDocument(document)),
    vscode.workspace.onDidSaveTextDocument((document) => refreshDiagnosticsForDocument(document)),
    vscode.workspace.onDidCloseTextDocument((document) => {
      if (isBlackLangSourceDocument(document)) {
        diagnosticCollection.delete(document.uri);
      }
    })
  );

  refreshDiagnosticsForOpenDocuments();
}

function deactivate() {
  if (diagnosticCollection) {
    diagnosticCollection.dispose();
    diagnosticCollection = null;
  }
}

async function refreshDiagnosticsForOpenDocuments() {
  if (!vscode || !diagnosticsEnabled()) {
    return;
  }
  await Promise.all(vscode.workspace.textDocuments.filter(isBlackLangSourceDocument).map(refreshDiagnosticsForDocument));
}

async function refreshDiagnosticsForPath(filePath) {
  if (!vscode) {
    return;
  }
  const document = vscode.workspace.textDocuments.find((candidate) => candidate.fileName === filePath);
  if (document) {
    await refreshDiagnosticsForDocument(document);
  }
}

async function refreshDiagnosticsForDocument(document) {
  if (!vscode || !diagnosticCollection || !diagnosticsEnabled() || !isBlackLangSourceDocument(document) || document.isUntitled) {
    return;
  }
  try {
    const payload = await runBlackJSON(["ide", "diagnostics", document.fileName, "--json"], workspaceFolderFor(document));
    diagnosticCollection.set(document.uri, vscodeDiagnosticsFromResult(payload));
  } catch (error) {
    const diagnostic = new vscode.Diagnostic(
      new vscode.Range(0, 0, 0, 1),
      `BlackLang diagnostics failed: ${error.message}`,
      vscode.DiagnosticSeverity.Error
    );
    diagnostic.source = "blacklang";
    diagnostic.code = "IDE_DIAGNOSTICS_FAILED";
    diagnosticCollection.set(document.uri, [diagnostic]);
  }
}

function completionItemsFromManifest(manifest) {
  const items = Array.isArray(manifest?.completionItems) ? manifest.completionItems : [];
  return items.map((item) => {
    const completion = new vscode.CompletionItem(item.label, completionKind(item.kind));
    completion.detail = item.detail || item.context;
    completion.insertText = item.insertText || item.label;
    completion.documentation = item.documentation ? new vscode.MarkdownString(item.documentation) : undefined;
    return completion;
  });
}

function snippetItemsFromManifest(manifest) {
  const snippets = Array.isArray(manifest?.snippets) ? manifest.snippets : [];
  return snippets.map((snippet) => {
    const completion = new vscode.CompletionItem(snippet.prefix, vscode.CompletionItemKind.Snippet);
    completion.detail = snippet.description;
    completion.insertText = new vscode.SnippetString((snippet.body || []).join("\n"));
    completion.documentation = snippet.description ? new vscode.MarkdownString(snippet.description) : undefined;
    return completion;
  });
}

function vscodeDiagnosticsFromResult(payload) {
  const diagnostics = Array.isArray(payload?.diagnostics) ? payload.diagnostics : [];
  return diagnostics.map((entry) => {
    const diagnostic = new vscode.Diagnostic(
      new vscode.Range(
        entry.range.start.line,
        entry.range.start.character,
        entry.range.end.line,
        entry.range.end.character
      ),
      entry.message,
      diagnosticSeverity(entry.severity)
    );
    diagnostic.source = entry.source ? `blacklang:${entry.source}` : "blacklang";
    diagnostic.code = entry.code;
    diagnostic.suggestion = entry.suggestion;
    return diagnostic;
  });
}

function codeActionsForDiagnostics(document, diagnostics) {
  const actions = [];
  let hasFormatFinding = false;
  for (const diagnostic of diagnostics) {
    if (diagnostic.code === "FORMAT_REQUIRED") {
      hasFormatFinding = true;
      const action = new vscode.CodeAction("Run black format", vscode.CodeActionKind.QuickFix);
      action.command = {
        title: "Run black format",
        command: "blacklang.formatSource",
        arguments: [document.uri],
      };
      action.diagnostics = [diagnostic];
      actions.push(action);
    }

    if (diagnostic.code) {
      const action = new vscode.CodeAction(`Show BlackLang diagnostic ${diagnostic.code}`, vscode.CodeActionKind.QuickFix);
      action.command = {
        title: `Show ${diagnostic.code}`,
        command: "blacklang.showDiagnosticInfo",
        arguments: [diagnostic],
      };
      action.diagnostics = [diagnostic];
      actions.push(action);
    }
  }

  if (hasFormatFinding) {
    const fixAll = new vscode.CodeAction("Run black format for this file", vscode.CodeActionKind.SourceFixAll);
    fixAll.command = {
      title: "Run black format for this file",
      command: "blacklang.formatSource",
      arguments: [document.uri],
    };
    actions.push(fixAll);
  }
  return actions;
}

function completionKind(kind) {
  switch (kind) {
    case "type":
      return vscode.CompletionItemKind.TypeParameter;
    case "modifier":
      return vscode.CompletionItemKind.Property;
    case "operator":
      return vscode.CompletionItemKind.Operator;
    case "value":
      return vscode.CompletionItemKind.Value;
    default:
      return vscode.CompletionItemKind.Keyword;
  }
}

function diagnosticSeverity(severity) {
  switch (severity) {
    case "warning":
      return vscode.DiagnosticSeverity.Warning;
    case "information":
      return vscode.DiagnosticSeverity.Information;
    case "hint":
      return vscode.DiagnosticSeverity.Hint;
    default:
      return vscode.DiagnosticSeverity.Error;
  }
}

function runBlackJSON(args, cwd) {
  return new Promise((resolve, reject) => {
    const cliPath = blackCLIPath();
    childProcess.execFile(cliPath, args, { cwd, windowsHide: true, maxBuffer: 10 * 1024 * 1024 }, (error, stdout, stderr) => {
      const text = stdout.trim();
      if (text.length > 0) {
        try {
          resolve(JSON.parse(text));
          return;
        } catch (parseError) {
          reject(new Error(`Could not parse BlackLang JSON: ${parseError.message}`));
          return;
        }
      }
      if (error) {
        reject(new Error(stderr.trim() || error.message));
        return;
      }
      reject(new Error("BlackLang CLI returned empty output."));
    });
  });
}

function blackCLIPath() {
  if (!vscode) {
    return "black";
  }
  return vscode.workspace.getConfiguration("blacklang").get("cliPath") || "black";
}

function diagnosticsEnabled() {
  if (!vscode) {
    return false;
  }
  return vscode.workspace.getConfiguration("blacklang").get("diagnostics.enable") !== false;
}

function workspaceFolderFor(document) {
  if (!vscode || !document) {
    return undefined;
  }
  return vscode.workspace.getWorkspaceFolder(document.uri)?.uri.fsPath;
}

function workspaceFolderForPath(filePath) {
  if (!vscode || !filePath) {
    return undefined;
  }
  const uri = vscode.Uri.file(filePath);
  return vscode.workspace.getWorkspaceFolder(uri)?.uri.fsPath ?? path.dirname(filePath);
}

function isBlackLangSourceDocument(document) {
  return Boolean(document && (document.languageId === "blacklang" || document.fileName.endsWith(".black")));
}

function symbolAtSelection(editor) {
  const range = editor.document.getWordRangeAtPosition(editor.selection.active, /[A-Za-z_][A-Za-z0-9_.]*/);
  if (!range) {
    return "";
  }
  return editor.document.getText(range);
}

async function showJSONDocument(title, payload) {
  const document = await vscode.workspace.openTextDocument({
    language: "json",
    content: JSON.stringify(payload, null, 2),
  });
  await vscode.window.showTextDocument(document, { preview: true });
  vscode.window.setStatusBarMessage(title, 3000);
}

module.exports = {
  activate,
  deactivate,
  __test: {
    completionKind,
    diagnosticSeverity,
    isBlackLangSourceDocument,
    vscodeDiagnosticsFromResult,
  },
};
