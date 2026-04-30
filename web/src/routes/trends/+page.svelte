<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type TrendsResponse } from '$lib/api';

	let trends: TrendsResponse | null = null;
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			trends = await api.trends();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load trends';
		} finally {
			loading = false;
		}
	});

	function formatAmount(amount: number) {
		return '$' + Math.abs(amount).toFixed(2);
	}

	function formatPct(current: number, previous: number): string {
		if (previous === 0) return '—';
		const pct = ((Math.abs(current) - Math.abs(previous)) / Math.abs(previous)) * 100;
		return (pct > 0 ? '+' : '') + pct.toFixed(0) + '%';
	}

	function pctClass(current: number, previous: number): string {
		if (previous === 0) return '';
		return Math.abs(current) > Math.abs(previous) ? 'worse' : 'better';
	}
</script>

<svelte:head>
	<title>Trends — miser</title>
</svelte:head>

<div>
	<div class="header">
		<h1>Trends</h1>
		{#if trends}
			<span class="period">{trends.previous_month} → {trends.current_month}</span>
		{/if}
	</div>

	{#if loading}
		<p class="state">Loading...</p>
	{:else if error}
		<p class="state error">{error}</p>
	{:else if trends}
		<table>
			<thead>
				<tr>
					<th>Category</th>
					<th class="right">This Month</th>
					<th class="right">Last Month</th>
					<th class="right">Change</th>
					<th class="right">Budget</th>
					<th>Status</th>
				</tr>
			</thead>
			<tbody>
				{#each trends.categories as cat}
					{@const budget = cat.budget ?? 0}
					{@const over = cat.pacing === 'over'}
					<tr class:parent={cat.subcategories && cat.subcategories.length > 0}>
						<td>
						<div class="cat-label">{cat.category}</div>
						{#if budget > 0 && cat.budget_used_pct !== undefined}
							{@const barPct = Math.min(100, cat.budget_used_pct)}
							<div class="parent-bar">
								<div
									class="parent-bar-fill"
									class:over
									style="width: {barPct}%"
								></div>
							</div>
						{/if}
					</td>
						<td class="right mono">{formatAmount(cat.current)}</td>
						<td class="right mono muted">{formatAmount(cat.previous)}</td>
						<td class="right mono {pctClass(cat.current, cat.previous)}">{formatPct(cat.current, cat.previous)}</td>
						<td class="right mono muted">{budget > 0 ? formatAmount(budget) : '—'}</td>
						<td>
							{#if budget > 0}
								<span class="badge" class:over class:under={!over}>
									{over ? 'OVER' : 'under'}
								</span>
							{/if}
						</td>
					</tr>
					{#if cat.subcategories}
						{#each cat.subcategories as sub}
							<tr class="subrow">
								<td class="sub-name">{sub.category}</td>
								<td class="right mono">{formatAmount(sub.current)}</td>
								<td class="right mono muted">{formatAmount(sub.previous)}</td>
								<td class="right mono {pctClass(sub.current, sub.previous)}">{formatPct(sub.current, sub.previous)}</td>
								<td class="right mono muted">—</td>
								<td></td>
							</tr>
						{/each}
					{/if}
				{/each}
			</tbody>
		</table>
	{/if}
</div>

<style>
	.header {
		display: flex;
		align-items: baseline;
		gap: 16px;
		margin-bottom: 32px;
	}

	h1 {
		font-family: var(--font-display);
		font-size: 2.118rem;
		font-weight: 600;
		color: var(--color-text);
		letter-spacing: -0.3px;
	}

	.period {
		font-size: 0.765rem;
		color: var(--color-text-muted);
	}

	.state {
		color: var(--color-text-muted);
		margin-top: 32px;
	}

	.error {
		color: var(--color-expense);
	}

	/* ── Table ────────────────────────────────────────── */
	table {
		width: 100%;
		border-collapse: collapse;
	}

	th {
		text-align: left;
		font-size: 0.647rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--color-text-muted);
		padding: 10px 14px;
		border-bottom: 1px solid var(--color-border);
		background: var(--color-bg);
	}

	td {
		padding: 10px 14px;
		border-bottom: 1px solid var(--color-border);
		font-size: 0.824rem;
		vertical-align: middle;
	}

	tbody tr:last-child td {
		border-bottom: none;
	}

	/* ── Parent rows ──────────────────────────────────── */
	tr.parent td {
		padding-top: 14px;
		padding-bottom: 10px;
	}

	.cat-label {
		font-family: var(--font-display);
		font-size: 1.059rem;
		font-weight: 600;
		color: var(--color-text);
		letter-spacing: -0.2px;
		line-height: 1.2;
		margin-bottom: 6px;
	}

	.parent-bar {
		height: 5px;
		background: var(--color-surface-alt);
		border-radius: 3px;
		overflow: hidden;
		width: 120px;
	}

	.parent-bar-fill {
		height: 100%;
		background: var(--color-accent);
		border-radius: 3px;
		transition: width 0.35s ease;
	}

	.parent-bar-fill.over {
		background: var(--color-expense);
	}

	/* ── Subrows ──────────────────────────────────────── */
	tr.subrow td {
		font-size: 0.765rem;
		color: var(--color-text-muted);
		padding-top: 6px;
		padding-bottom: 6px;
		background: var(--color-surface-alt);
	}

	.sub-name {
		padding-left: 38px;
	}

	/* ── Utilities ────────────────────────────────────── */
	.right {
		text-align: right;
	}

	.mono {
		font-family: var(--font-mono);
		font-size: 0.765rem;
	}

	.muted {
		color: var(--color-text-muted);
	}

	.better {
		color: var(--color-income);
	}

	.worse {
		color: var(--color-expense);
	}

	/* ── Badges ───────────────────────────────────────── */
	.badge {
		display: inline-block;
		font-size: 0.647rem;
		font-weight: 500;
		padding: 3px 8px;
		border-radius: 100px;
		letter-spacing: 0.06em;
	}

	.badge.over {
		background: color-mix(in srgb, var(--color-expense) 12%, transparent);
		color: var(--color-expense);
	}

	.badge.under {
		background: color-mix(in srgb, var(--color-income) 12%, transparent);
		color: var(--color-income);
	}
</style>
