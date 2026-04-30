<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type TrendsResponse, type CategoryTotal } from '$lib/api';

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

	function pctChange(current: number, previous: number): string {
		if (previous === 0) return '—';
		const pct = ((Math.abs(current) - Math.abs(previous)) / Math.abs(previous)) * 100;
		return (pct > 0 ? '+' : '') + pct.toFixed(0) + '%';
	}

	function pctClass(current: number, previous: number): string {
		if (previous === 0) return '';
		return Math.abs(current) > Math.abs(previous) ? 'worse' : 'better';
	}

	$: budgetMap = Object.fromEntries(
		(trends?.budgets ?? []).map((b) => [b.category, b.budget])
	);

	$: previousMap = Object.fromEntries(
		(trends?.previous ?? []).map((c) => [c.category, c.total])
	);
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
				{#each trends.current as cat}
					{@const prev = previousMap[cat.category] ?? 0}
					{@const budget = budgetMap[cat.category] ?? 0}
					{@const over = budget > 0 && Math.abs(cat.total) > budget}
					<tr class:parent={cat.subcategories && cat.subcategories.length > 0}>
						<td>
						<div class="cat-label">{cat.category}</div>
						{#if budgetMap[cat.category] > 0}
							{@const catPct = Math.min(100, (Math.abs(cat.total) / budgetMap[cat.category]) * 100)}
							<div class="parent-bar">
								<div
									class="parent-bar-fill"
									class:over={catPct >= 100}
									style="width: {catPct}%"
								></div>
							</div>
						{/if}
					</td>
						<td class="right mono">{formatAmount(cat.total)}</td>
						<td class="right mono muted">{formatAmount(prev)}</td>
						<td class="right mono {pctClass(cat.total, prev)}">{pctChange(cat.total, prev)}</td>
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
							{@const subPrev = previousMap[sub.category] ?? 0}
							<tr class="subrow">
								<td class="sub-name">{sub.category}</td>
								<td class="right mono">{formatAmount(sub.total)}</td>
								<td class="right mono muted">{formatAmount(subPrev)}</td>
								<td class="right mono {pctClass(sub.total, subPrev)}">{pctChange(sub.total, subPrev)}</td>
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
