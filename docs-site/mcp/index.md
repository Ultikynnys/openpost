# Assistant Scheduling With MCP

OpenPost's MCP support lets ChatGPT-style clients and local desktop assistants work with your scheduler through the same authenticated OpenPost instance you use in the web app and CLI.

Use it when you want an assistant to:

- inspect workspaces, connected accounts, media, drafts, providers, and scheduled posts
- turn a rough idea into a draft, then adapt that draft for each destination
- adapt copy for each destination before scheduling
- attach existing workspace media or upload media from a public URL
- suggest the next posting slot, schedule approved posts, or cancel queued posts

## Ways to connect

### ChatGPT-style clients

Use the remote MCP endpoint from your OpenPost instance:

```txt
https://your-openpost-host.example/mcp
```

OAuth-aware clients can use OpenPost's browser account-linking flow. Clients that need a manual token can use a dedicated `mcp:full` token from **Settings -> Account -> CLI Devices & API Tokens**.

When approving OAuth or creating a manual token, prefer the current-workspace boundary unless the client truly needs every workspace you can access.

### Desktop MCP clients

Install and authenticate the OpenPost CLI with the MCP proxy, then run the local stdio proxy:

```sh
curl -fsSL https://raw.githubusercontent.com/rodrgds/openpost/main/scripts/install-cli.sh | sh -s -- --with-mcp
openpost --profile local auth login https://your-openpost-host.example
openpost-mcp --profile local
```

The proxy reads the selected CLI profile and forwards MCP frames to the remote `/mcp` endpoint. It does not open the database and does not need provider secrets on the client machine.

## Current assistant tools

OpenPost advertises a compact four-tool surface so connecting it does not load
every scheduling schema into the assistant's context:

- `search_operations` finds relevant OpenPost operations and returns only the
  schemas needed for the current task, including whether each result must use
  `query_operation` or `execute_operation`. It returns no result instead of
  guessing when a request is ambiguous or outside OpenPost.
- `query_operation` runs only guaranteed read-only operations and rejects
  mutations.
- `execute_operation` runs only state-changing or external-action operations
  and rejects read-only work, giving clients a hard approval boundary.
- `render_scheduler_widget` displays a visual scheduler summary in clients that
  support MCP Apps resources.

The discoverable operations still cover workspaces, providers, accounts, media,
drafts, renditions, format-first publications, validation, scheduling,
publishing, status, cancellation, lifecycle events, comments, and slot
suggestions. You can usually ask for the outcome in plain language; the
assistant should use `search_operations` before `query_operation` or
`execute_operation` when it needs an operation schema.

## Safe workflow

1. Ask the assistant to inspect the current workspace, provider catalog, accounts, and recent media.
2. Draft the base post and destination-specific renditions.
3. Review the proposed schedule and destination list.
4. Let the assistant schedule only after the final content and accounts are correct.

MCP tools validate workspace membership, optional token workspace boundaries, and account ownership before reading or changing data. Scheduling and media uploads use the same quota and usage accounting as the web app and CLI.

## Activity and revocation

Recent MCP tool calls appear in **Settings -> Account -> CLI Devices & API Tokens** with client attribution when the request used a dedicated MCP or CLI token. Revoke the token there to disconnect a client.

For protocol details, Apps SDK metadata, OAuth discovery, and implementation notes, see [MCP And ChatGPT App](/development/mcp) in the developer docs.
