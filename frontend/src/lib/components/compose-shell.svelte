<script lang="ts">
	import { page } from '$app/state';
	import { replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ComposeSimple from './compose-simple.svelte';
	import ComposeFocusedPublication from './compose-focused-publication.svelte';
	import ComposeModeSelect from './compose-mode-select.svelte';
	import { ui } from '$lib/stores/ui.svelte';
	import { isLegacyComposerMode, type ComposerModeKey } from './compose/modes';

	let selectedMode = $state<ComposerModeKey>('short_text');
	let lastComposerThreadState = false;
	const initialScheduleDate = $derived(page.url.searchParams.get('date'));
	const initialWorkspaceId = $derived(page.url.searchParams.get('workspace_id'));
	const composerResetCounter = $derived(ui.composerResetCounter);

	function handleThreadStateChange(isThread: boolean) {
		if (isThread) {
			selectedMode = 'thread';
		} else if (lastComposerThreadState && selectedMode === 'thread') {
			selectedMode = 'short_text';
		}
		lastComposerThreadState = isThread;
	}

	function handleDraftCreated(id: string) {
		ui.setActiveComposerDraft(id);
		replaceState(resolve(`/posts/${id}` as '/'), {});
	}

	function handleComposerReset() {
		ui.clearActiveComposerDraft();
		replaceState(resolve('/'), {});
	}
</script>

<div class="flex min-h-0 flex-1 flex-col bg-background" data-testid="compose-shell">
	{#key composerResetCounter}
		{#if isLegacyComposerMode(selectedMode)}
			<div data-testid="legacy-composer-shell" class="flex min-h-0 flex-1 flex-col">
				<ComposeSimple
					{initialScheduleDate}
					{initialWorkspaceId}
					onDraftCreated={handleDraftCreated}
					onDeleted={handleComposerReset}
					onSuccess={handleComposerReset}
					onThreadStateChange={handleThreadStateChange}
				>
					{#snippet modeControl()}
						<ComposeModeSelect
							{selectedMode}
							compactOnNarrow
							onModeChange={(mode) => (selectedMode = mode)}
						/>
					{/snippet}
				</ComposeSimple>
			</div>
		{:else}
			{#key selectedMode}
				<ComposeFocusedPublication mode={selectedMode}>
					{#snippet modeControl()}
						<ComposeModeSelect
							{selectedMode}
							compactOnNarrow
							onModeChange={(mode) => (selectedMode = mode)}
						/>
					{/snippet}
				</ComposeFocusedPublication>
			{/key}
		{/if}
	{/key}
</div>
