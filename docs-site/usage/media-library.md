# Media Library

OpenPost stores uploaded media in the selected workspace and shows it in the Media Library. The library is the place to inspect assets, favorite them, check usage, upload new files, and clean up media that is no longer needed.

## Basic flow

1. Upload media from the Media Library or while composing a post.
2. Attach media to posts through the composer, CLI, API, or MCP tools.
3. Reuse existing media by ID from automation surfaces such as the CLI/API/MCP. The current web composer focuses on uploading files for the active draft; it does not yet expose a full “pick existing media from library” dialog.

## Operational note

Do not remove media that is still needed by scheduled posts, especially when Threads will need a public media URL at publish time.
