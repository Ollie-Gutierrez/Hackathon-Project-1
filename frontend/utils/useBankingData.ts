import { useState, useEffect, useCallback } from 'react';
import { getStoredTokens } from './auth';
import { liminalApi } from './liminalApi';
import { generateMockBankingData } from './mockBankingData';

export interface BankingData {
  balance: {
    wallet: number;
    savings: number;
    total: number;
    currency: string;
  };
  transactions: Array<{
    id: string;
    amount: number;
    currency: string;
    type: 'send' | 'receive' | 'deposit' | 'withdrawal';
    status: 'pending' | 'completed' | 'failed';
    description: string;
    createdAt: string;
    counterparty?: string;
  }>;
  profile: {
    id: string;
    email: string;
    name?: string;
    createdAt: string;
    verified: boolean;
  };
  spending: Array<{
    category: string;
    amount: number;
    percentage: number;
    count: number;
  }>;
  savingsRate: number | null;
}

// Toggle this to switch between mock and real data
const USE_MOCK_DATA = true; // Set to false when you have real data

export function useBankingData() {
  const [data, setData] = useState<BankingData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    // If using mock data, skip API calls
    if (USE_MOCK_DATA) {
      setIsLoading(true);
      // Simulate network delay
      await new Promise(resolve => setTimeout(resolve, 800));
      setData(generateMockBankingData());
      setIsLoading(false);
      return;
    }

    // Real API calls
    const tokens = getStoredTokens();
    if (!tokens?.accessToken) {
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const [balanceData, savingsData, ratesData, transactionsData, profileData] = await Promise.all([
        liminalApi.getBalance(tokens.accessToken),
        liminalApi.getSavingsBalance(tokens.accessToken),
        liminalApi.getVaultRates(tokens.accessToken),
        liminalApi.getTransactions(tokens.accessToken, 1, 10),
        liminalApi.getProfile(tokens.accessToken),
      ]);

      const walletBalance = balanceData.balance;
      const savingsBalance = savingsData.balance;
      const savingsRate = ratesData.rates[0]?.apy || null;

      // Calculate real spending from transactions
      const spending = calculateSpending(transactionsData.transactions);

      setData({
        balance: {
          wallet: walletBalance,
          savings: savingsBalance,
          total: walletBalance + savingsBalance,
          currency: balanceData.currency,
        },
        transactions: transactionsData.transactions,
        profile: profileData,
        spending,
        savingsRate,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch banking data');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return {
    balance: data?.balance || { wallet: 0, savings: 0, total: 0, currency: 'USD' },
    transactions: data?.transactions || [],
    profile: data?.profile,
    spending: data?.spending || [],
    savingsRate: data?.savingsRate || null,
    isLoading,
    error,
    refetch: fetchData,
  };
}

/**
 * Calculate spending categories from transactions
 */
function calculateSpending(transactions: BankingData['transactions']): BankingData['spending'] {
  // Filter for outgoing transactions only
  const outgoing = transactions.filter(tx =>
    (tx.type === 'send' || tx.type === 'withdrawal') && tx.status === 'completed'
  );

  if (outgoing.length === 0) {
    return [];
  }

  // Group by category
  const categoryMap: Record<string, { amount: number; count: number }> = {};
  let total = 0;

  outgoing.forEach(tx => {
    const category = categorizeTransaction(tx);
    if (!categoryMap[category]) {
      categoryMap[category] = { amount: 0, count: 0 };
    }
    categoryMap[category].amount += tx.amount;
    categoryMap[category].count += 1;
    total += tx.amount;
  });

  // Convert to array with percentages
  const categories = Object.entries(categoryMap)
    .map(([category, data]) => ({
      category,
      amount: Math.round(data.amount * 100) / 100,
      count: data.count,
      percentage: total > 0 ? Math.round((data.amount / total) * 100) : 0
    }))
    .sort((a, b) => b.amount - a.amount)
    .slice(0, 5); // Top 5 categories

  return categories;
}

/**
 * Categorize a transaction based on description and counterparty
 */
function categorizeTransaction(tx: BankingData['transactions'][0]): string {
  const text = `${tx.description} ${tx.counterparty || ''}`.toLowerCase();

  // Food & Dining
  if (text.match(/food|restaurant|dining|cafe|coffee|lunch|dinner|breakfast|pizza|burger|sushi|delivery|uber eats|doordash|grubhub/)) {
    return 'Food & Dining';
  }

  // Transportation
  if (text.match(/uber|lyft|taxi|transport|gas|fuel|parking|transit|metro|bus|train|ride/)) {
    return 'Transportation';
  }

  // Shopping
  if (text.match(/amazon|shopping|store|retail|mall|purchase|buy|walmart|target|ebay/)) {
    return 'Shopping';
  }

  // Entertainment
  if (text.match(/entertainment|movie|cinema|netflix|spotify|hulu|disney|hbo|music|concert|game|theater|steam|playstation/)) {
    return 'Entertainment';
  }

  // Bills & Utilities
  if (text.match(/bill|utility|electric|water|internet|phone|subscription|insurance|rent|mortgage/)) {
    return 'Bills & Utilities';
  }

  // Healthcare
  if (text.match(/health|medical|doctor|pharmacy|hospital|dental|clinic/)) {
    return 'Healthcare';
  }

  // Transfer/Savings
  if (text.match(/transfer|saving|deposit|withdraw/)) {
    return 'Transfers';
  }

  return 'Other';
}