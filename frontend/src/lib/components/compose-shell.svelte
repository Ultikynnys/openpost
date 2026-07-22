<script lang="ts">
	import { page } from '$app/state';
	import { goto, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ComposeSimple from './compose-simple.svelte';
	import ComposeFocusedPublication from './compose-focused-publication.svelte';
	import ComposeModeSelect from './compose-mode-select.svelte';
	import SampleCampaign from './sample-campaign.svelte';
	import { ui } from '$lib/stores/ui.svelte';
	import { isLegacyComposerMode, type ComposerModeKey } from './compose/modes';
	import { hostedPlanFromSearchParams, settingsPathForPlan } from '$lib/billing';
	import { isSampleCampaignRequested, SAMPLE_CAMPAIGN_DISMISSED_KEY } from '$lib/sample-campaign';
	import { m } from '$lib/paraglide/messages';

	let selectedMode = $state<ComposerModeKey>('short_text');
	let lastComposerThreadState = false;
	const initialScheduleDate = $derived(page.url.searchParams.get('date'));
	const initialWorkspaceId = $derived(page.url.searchParams.get('workspace_id'));
	const composerResetCounter = $derived(ui.composerResetCounter);
	const sampleCampaignActive = $derived(isSampleCampaignRequested(page.url.searchParams));
	const sampleCampaignPlan = $derived(hostedPlanFromSearchParams(page.url.searchParams));
	const sampleContinueLabel = $derived(
		sampleCampaignPlan
			? m.sample_campaign_continue_checkout()
			: m.sample_campaign_continue_accounts()
	);

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

	function rememberSampleCampaignChoice() {
		localStorage.setItem(SAMPLE_CAMPAIGN_DISMISSED_KEY, 'true');
	}

	async function skipSampleCampaign() {
		rememberSampleCampaignChoice();
		if (sampleCampaignPlan) {
			await goto(resolve(settingsPathForPlan(sampleCampaignPlan) as '/'));
			return;
		}
		replaceState(resolve('/'), {});
	}

	async function continueFromSampleCampaign() {
		rememberSampleCampaignChoice();
		const target = sampleCampaignPlan ? settingsPathForPlan(sampleCampaignPlan) : '/accounts';
		await goto(resolve(target as '/'));
	}
</script>

<div class="flex min-h-0 flex-1 flex-col bg-background" data-testid="compose-shell">
	{#if sampleCampaignActive}
		<SampleCampaign
			onSkip={skipSampleCampaign}
			onContinue={continueFromSampleCampaign}
			continueLabel={sampleContinueLabel}
		/>
	{:else}
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
	{/if}
</div>
