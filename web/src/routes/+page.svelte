<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type TrendsResponse, type Transaction, type Report } from '$lib/api';

	let trends: TrendsResponse | null = null;
	let recentTxns: Transaction[] = [];
	let report: Report | null = null;
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			[trends, recentTxns, report] = await Promise.all([
				api.trends(),
				api.transactions({ limit: '10' }),
				api.latestReport()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load dashboard';
		} finally {
			loading = false;
		}
	});

	function formatAmount(amount: number) {
		const abs = Math.abs(amount);
		return (amount < 0 ? '-' : '+') + '$' + abs.toFixed(2);
	}

	function amountClass(amount: number) {
		return amount < 0 ? 'expense' : 'income';
	}

	$: topCategories = trends?.current.slice(0, 5) ?? [];
	$: budgetMap = Object.fromEntries(
		(trends?.budgets ?? []).map((b) => [b.category, b.budget])
	);
	$: heroTotal =
		trends?.current
			.filter((cat) => cat.total < 0)
			.reduce((sum, cat) => sum + Math.abs(cat.total), 0) ?? 0;
</script>

<svelte:head>
	<title>Dashboard — miser</title>
</svelte:head>

<div class="dashboard">
	<header class="hero">
		<div class="hero-month">{trends?.current_month?.toUpperCase() ?? 'LOADING'}</div>
		<div class="hero-total">
			${heroTotal.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
		</div>
		<div class="hero-sub">
			{#if !loading && trends}
				Compared to last month
			{/if}
		</div>
	</header>
	<hr class="divider" />

	{#if loading}
		<p class="loading">Loading...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<div class="grid">
			<!-- Top spending categories -->
			<section class="card">
				<h2>Top Categories</h2>
				<ul class="category-list">
					{#each topCategories as cat}
						{@const budget = budgetMap[cat.category]}
						{@const pct = budget ? Math.min(100, (Math.abs(cat.total) / budget) * 100) : null}
						<li>
							<div class="cat-row">
								<span class="cat-name">{cat.category}</span>
								<span class="cat-amount {amountClass(cat.total)}">{formatAmount(cat.total)}</span>
							</div>
							{#if pct !== null}
								<div class="budget-bar-row">
									<div class="budget-bar">
										<div
											class="budget-fill"
											class:over={pct >= 100}
											style="width: {pct}%"
										></div>
									</div>
									<span class="budget-pct" class:over={pct >= 100}>{pct.toFixed(0)}%</span>
								</div>
							{/if}
						</li>
					{/each}
				</ul>
			</section>

			<!-- Recent transactions -->
			<section class="card">
				<h2>Recent Transactions</h2>
				<ul class="txn-list">
					{#each recentTxns as txn}
						<li class="txn-row">
							<div class="txn-left">
								<span class="txn-merchant">{txn.merchant_clean ?? txn.merchant}</span>
								<span class="txn-cat">{txn.category_name || 'Uncategorized'}</span>
							</div>
							<div class="txn-right">
								<span class="txn-amount {amountClass(txn.amount)}">{formatAmount(txn.amount)}</span>
								<span class="txn-date">{txn.date}</span>
							</div>
						</li>
					{/each}
				</ul>
				<a href="/transactions" class="view-all">View all transactions →</a>
			</section>

			<!-- Latest report narrative -->
			{#if report}
				<section class="card narrative">
					<h2>Monthly Report — {report.year}/{String(report.month).padStart(2, '0')}</h2>
					<div class="narrative-body">
						<p>{report.narrative}</p>
					</div>
				</section>
			{/if}
		</div>
	{/if}
</div>

<style>
	.dashboard {
		max-width: 1600px;
	}

	/* ── Hero ─────────────────────────────────────────── */
	.hero {
		margin-bottom: 32px;
	}

	.hero-month {
		font-size: 11px;
		font-weight: 500;
		letter-spacing: 0.14em;
		color: var(--color-accent);
		margin-bottom: 8px;
	}

	.hero-total {
		font-family: var(--font-display);
		font-size: 80px;
		font-weight: 600;
		color: var(--color-text);
		line-height: 1;
		letter-spacing: -1px;
	}

	.hero-sub {
		font-size: 15px;
		color: var(--color-text-muted);
		margin-top: 10px;
		min-height: 1.5em;
	}

	.divider {
		border: none;
		border-top: 1px solid var(--color-border);
		margin-bottom: 36px;
	}

	/* ── Loading / error states ───────────────────────── */
	.loading,
	.error {
		color: var(--color-text-muted);
		margin-top: 32px;
	}

	.error {
		color: var(--color-expense);
	}

	/* ── Grid ─────────────────────────────────────────── */
	.grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 24px;
	}

	.card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 24px;
	}

	/* ── Section headers ──────────────────────────────── */
	h2 {
		font-size: 11px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--color-text-muted);
		margin-bottom: 20px;
	}

	/* ── Categories ───────────────────────────────────── */
	.category-list {
		list-style: none;
		display: flex;
		flex-direction: column;
	}

	.category-list li {
		padding: 12px 0;
		border-bottom: 1px solid var(--color-border);
	}

	.category-list li:last-child {
		border-bottom: none;
	}

	.cat-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 7px;
	}

	.cat-name {
		font-size: 14px;
		font-weight: 500;
		color: var(--color-text);
	}

	.cat-amount {
		font-family: var(--font-mono);
		font-size: 14px;
	}

	.budget-bar-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.budget-bar {
		flex: 1;
		height: 6px;
		background: var(--color-surface-alt);
		border-radius: 3px;
		overflow: hidden;
	}

	.budget-fill {
		height: 100%;
		background: var(--color-accent);
		border-radius: 3px;
		transition: width 0.35s ease;
	}

	.budget-fill.over {
		background: var(--color-expense);
	}

	.budget-pct {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
		min-width: 32px;
		text-align: right;
	}

	.budget-pct.over {
		color: var(--color-expense);
	}

	/* ── Transactions ─────────────────────────────────── */
	.txn-list {
		list-style: none;
		display: flex;
		flex-direction: column;
	}

	.txn-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 11px 8px;
		margin: 0 -8px;
		border-bottom: 1px solid var(--color-border);
		border-radius: var(--radius);
		transition: background 0.1s;
	}

	.txn-row:hover {
		background: var(--color-surface-alt);
	}

	.txn-row:last-child {
		border-bottom: none;
	}

	.txn-left {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.txn-merchant {
		font-size: 15px;
		font-weight: 500;
		color: var(--color-text);
	}

	.txn-cat {
		font-size: 13px;
		color: var(--color-text-muted);
	}

	.txn-right {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 3px;
	}

	.txn-amount {
		font-family: var(--font-mono);
		font-size: 14px;
	}

	.txn-date {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.view-all {
		display: block;
		margin-top: 16px;
		font-size: 13px;
		color: var(--color-text-muted);
		text-align: right;
		transition: color 0.12s;
	}

	.view-all:hover {
		color: var(--color-accent);
	}

	/* ── Narrative ────────────────────────────────────── */
	.narrative {
		grid-column: 1 / -1;
	}

	.narrative-body {
		border-left: 2px solid var(--color-accent);
		padding-left: 20px;
		margin-top: 4px;
	}

	.narrative-body p {
		font-family: var(--font-display);
		font-style: italic;
		font-size: 17px;
		line-height: 1.75;
		color: var(--color-text-muted);
	}

	/* ── Amount colors ────────────────────────────────── */
	.income {
		color: var(--color-income);
	}

	.expense {
		color: var(--color-expense);
	}
</style>
