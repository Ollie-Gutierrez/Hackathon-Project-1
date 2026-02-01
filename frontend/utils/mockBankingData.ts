import type { BankingData } from './useBankingData';

/**
 * Generate realistic mock banking data for testing
 */
export function generateMockBankingData(): BankingData {
    return {
        balance: {
            wallet: 2847.50,
            savings: 15420.30,
            total: 18267.80,
            currency: 'USD'
        },
        transactions: generateMockTransactions(),
        profile: {
            id: 'user_mock_123',
            email: 'demo@liminal.cash',
            name: 'Demo User',
            createdAt: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString(), // 90 days ago
            verified: true
        },
        spending: [
            { category: 'Food & Dining', amount: 450.25, percentage: 35, count: 12 },
            { category: 'Transportation', amount: 280.50, percentage: 22, count: 8 },
            { category: 'Shopping', amount: 320.75, percentage: 25, count: 6 },
            { category: 'Entertainment', amount: 150.00, percentage: 12, count: 4 },
            { category: 'Bills & Utilities', amount: 80.00, percentage: 6, count: 3 },
        ],
        savingsRate: 4.5
    };
}

/**
 * Generate realistic mock transactions
 */
function generateMockTransactions(): BankingData['transactions'] {
    const now = Date.now();
    const dayInMs = 24 * 60 * 60 * 1000;

    const transactionTemplates = [
        // Food & Dining
        { description: 'Starbucks Coffee', amount: 8.50, type: 'send' as const, category: 'food' },
        { description: 'Chipotle Mexican Grill', amount: 15.75, type: 'send' as const, category: 'food' },
        { description: 'Whole Foods Market', amount: 67.30, type: 'send' as const, category: 'food' },
        { description: 'DoorDash - Pizza Delivery', amount: 32.50, type: 'send' as const, category: 'food' },
        { description: 'Local Coffee Shop', amount: 6.25, type: 'send' as const, category: 'food' },

        // Transportation
        { description: 'Uber Ride', amount: 18.50, type: 'send' as const, category: 'transport' },
        { description: 'Gas Station', amount: 45.00, type: 'send' as const, category: 'transport' },
        { description: 'Lyft Ride', amount: 22.75, type: 'send' as const, category: 'transport' },
        { description: 'Metro Card Reload', amount: 30.00, type: 'send' as const, category: 'transport' },

        // Shopping
        { description: 'Amazon.com', amount: 89.99, type: 'send' as const, category: 'shopping' },
        { description: 'Target Store', amount: 54.25, type: 'send' as const, category: 'shopping' },
        { description: 'Nike Store', amount: 125.00, type: 'send' as const, category: 'shopping' },

        // Entertainment
        { description: 'Netflix Subscription', amount: 15.99, type: 'send' as const, category: 'entertainment' },
        { description: 'Spotify Premium', amount: 10.99, type: 'send' as const, category: 'entertainment' },
        { description: 'Movie Theater', amount: 28.50, type: 'send' as const, category: 'entertainment' },
        { description: 'Steam Games', amount: 59.99, type: 'send' as const, category: 'entertainment' },

        // Bills & Utilities
        { description: 'Electric Bill Payment', amount: 125.50, type: 'send' as const, category: 'bills' },
        { description: 'Internet Service', amount: 79.99, type: 'send' as const, category: 'bills' },
        { description: 'Phone Bill', amount: 65.00, type: 'send' as const, category: 'bills' },

        // Income/Receives
        { description: 'Payroll Deposit', amount: 2500.00, type: 'receive' as const, category: 'income' },
        { description: 'Freelance Payment', amount: 450.00, type: 'receive' as const, category: 'income' },
        { description: 'Refund from Amazon', amount: 29.99, type: 'receive' as const, category: 'refund' },
        { description: 'Payment from @alice', amount: 75.00, type: 'receive' as const, category: 'p2p' },

        // Savings
        { description: 'Savings Deposit', amount: 200.00, type: 'deposit' as const, category: 'savings' },
        { description: 'Savings Withdrawal', amount: 100.00, type: 'withdrawal' as const, category: 'savings' },
    ];

    // Generate 20 transactions over the last 30 days
    const transactions: BankingData['transactions'] = [];

    for (let i = 0; i < 20; i++) {
        const template = transactionTemplates[Math.floor(Math.random() * transactionTemplates.length)];
        const daysAgo = Math.floor(Math.random() * 30);
        const createdAt = new Date(now - (daysAgo * dayInMs));

        // Add some variance to amounts
        const amountVariance = 0.8 + (Math.random() * 0.4); // 80% to 120%
        const amount = Math.round(template.amount * amountVariance * 100) / 100;

        transactions.push({
            id: `tx_mock_${i}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            amount,
            currency: 'USD',
            type: template.type,
            status: 'completed',
            description: template.description,
            createdAt: createdAt.toISOString(),
            counterparty: template.type === 'receive' && Math.random() > 0.5 ? '@alice' : undefined
        });
    }

    // Sort by date (most recent first)
    transactions.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());

    return transactions;
}

/**
 * Generate mock data with custom parameters
 */
export function generateCustomMockData(options?: {
    walletBalance?: number;
    savingsBalance?: number;
    transactionCount?: number;
    savingsRate?: number;
}): BankingData {
    const base = generateMockBankingData();

    if (options?.walletBalance !== undefined) {
        base.balance.wallet = options.walletBalance;
    }

    if (options?.savingsBalance !== undefined) {
        base.balance.savings = options.savingsBalance;
    }

    base.balance.total = base.balance.wallet + base.balance.savings;

    if (options?.savingsRate !== undefined) {
        base.savingsRate = options.savingsRate;
    }

    if (options?.transactionCount !== undefined && options.transactionCount !== 20) {
        const allTxs = generateMockTransactions();
        base.transactions = allTxs.slice(0, options.transactionCount);
    }

    return base;
}