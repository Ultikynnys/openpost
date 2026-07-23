<script lang="ts">
	import { page } from '$app/state';
	import { goto, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ComposeFocusedPublication from './compose-focused-publication.svelte';
	import ComposeModeSelect from './compose-mode-select.svelte';
	import SampleCampaign from './sample-campaign.svelte';
	import { ui } from '$lib/stores/ui.svelte';
	import { type ComposerModeKey } from './compose/modes';
	import { hostedPlanFromSearchParams, settingsPathForPlan } from '$lib/billing';
	import { isSampleCampaignRequested, SAMPLE_CAMPAIGN_DISMISSED_KEY } from '$lib/sample-campaign';
	import { m } from '$lib/paraglide/messages';

	let selectedMode = $state<ComposerModeKey>('post');
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
			{#key selectedMode}
				<ComposeFocusedPublication
					mode={selectedMode}
					{initialScheduleDate}
					{initialWorkspaceId}
					onSuccess={handleComposerReset}
				>
					{#snippet modeControl()}
						<ComposeModeSelect
							{selectedMode}
							compactOnNarrow
							onModeChange={(mode) => (selectedMode = mode)}
						/>
					{/snippet}
				</ComposeFocusedPublication>
			{/key}
		{/key}
	{/if}
</div>
