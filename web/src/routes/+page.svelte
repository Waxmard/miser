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
</script>

<svelte:head>
	<title>Dashboard — miser</title>
</svelte:head>

<div class="dashboard">
	<h1>Dashboard</h1>
	<p class="subtitle">{trends?.current_month ?? '...'}</p>

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
								<div class="budget-bar">
									<div
										class="budget-fill"
										class:over={pct >= 100}
										style="width: {pct}%"
									></div>
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
			</section>

			<!-- Latest report narrative -->
			{#if report}
				<section class="card narrative">
					<h2>Monthly Report — {report.year}/{String(report.month).padStart(2, '0')}</h2>
					<p>{report.narrative}</p>
				</section>
			{/if}
		</div>
	{/if}
</div>

<style>
	.dashboard {
		max-width: 1100px;
	}

	h1 {
		font-size: 24px;
		font-weight: 700;
		margin-bottom: 4px;
	}

	.subtitle {
		color: var(--color-text-muted);
		margin-bottom: 32px;
	}

	.loading,
	.error {
		color: var(--color-text-muted);
		margin-top: 32px;
	}

	.error {
		color: var(--color-expense);
	}

	.grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 20px;
	}

	.card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 20px;
	}

	.narrative {
		grid-column: 1 / -1;
		color: var(--color-text-muted);
		line-height: 1.7;
	}

	h2 {
		font-size: 13px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-text-muted);
		margin-bottom: 16px;
	}

	/* Categories */
	.category-list {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.cat-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 4px;
	}

	.cat-name {
		font-size: 14px;
	}

	.cat-amount {
		font-family: var(--font-mono);
		font-size: 13px;
	}

	.budget-bar {
		height: 3px;
		background: var(--color-border);
		border-radius: 2px;
		overflow: hidden;
	}

	.budget-fill {
		height: 100%;
		background: var(--color-accent);
		border-radius: 2px;
		transition: width 0.3s ease;
	}

	.budget-fill.over {
		background: var(--color-expense);
	}

	/* Transactions */
	.txn-list {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.txn-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px 0;
		border-bottom: 1px solid var(--color-border);
	}

	.txn-row:last-child {
		border-bottom: none;
	}

	.txn-left {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.txn-merchant {
		font-size: 14px;
	}

	.txn-cat {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.txn-right {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 2px;
	}

	.txn-amount {
		font-family: var(--font-mono);
		font-size: 13px;
	}

	.txn-date {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.income {
		color: var(--color-income);
	}

	.expense {
		color: var(--color-expense);
	}
</style>
