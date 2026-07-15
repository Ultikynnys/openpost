<script lang="ts">
	import { page } from '$app/state';
	import ComposeSimple from './compose-simple.svelte';
	import ComposeFocusedPublication from './compose-focused-publication.svelte';
	import ComposeModeSelect from './compose-mode-select.svelte';
	import { isLegacyComposerMode, type ComposerModeKey } from './compose/modes';

	let selectedMode = $state<ComposerModeKey>('short_text');
	let lastComposerThreadState = false;
	const initialScheduleDate = $derived(page.url.searchParams.get('date'));
	const initialWorkspaceId = $derived(page.url.searchParams.get('workspace_id'));

	function handleThreadStateChange(isThread: boolean) {
		if (isThread) {
			selectedMode = 'thread';
		} else if (lastComposerThreadState && selectedMode === 'thread') {
			selectedMode = 'short_text';
		}
		lastComposerThreadState = isThread;
	}
</script>

<div class="flex min-h-0 flex-1 flex-col bg-background" data-testid="compose-shell">
	{#if isLegacyComposerMode(selectedMode)}
		<div data-testid="legacy-composer-shell" class="flex min-h-0 flex-1 flex-col">
			<ComposeSimple
				{initialScheduleDate}
				{initialWorkspaceId}
				onThreadStateChange={handleThreadStateChange}
			>
				{#snippet modeControl()}
					<ComposeModeSelect {selectedMode} onModeChange={(mode) => (selectedMode = mode)} />
				{/snippet}
			</ComposeSimple>
		</div>
	{:else}
		{#key selectedMode}
			<ComposeFocusedPublication mode={selectedMode}>
				{#snippet modeControl()}
					<ComposeModeSelect {selectedMode} onModeChange={(mode) => (selectedMode = mode)} />
				{/snippet}
			</ComposeFocusedPublication>
		{/key}
	{/if}
</div>
