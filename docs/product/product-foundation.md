# Savio — Product Foundation

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Business Requirements](business-requirements.md) — detailed business rules derived from this vision.
- [User Flows](user-flows.md) — end-to-end user workflows.
- [AI Architecture](../architecture/ai-architecture.md) — how AI interprets the product's financial facts.
- [Design System](../../DESIGN.md) — visual identity and UX guidelines.

## 1. Product Overview

**Savio** is an AI-powered personal cashflow intelligence and financial decision support platform.

Savio helps users:

- understand their current financial condition,
- track and analyze income and expenses,
- recognize spending patterns,
- forecast future cashflow,
- simulate financial decisions before taking action,
- and receive AI-assisted explanations based on their financial data.

Savio is not designed merely as an expense tracker.

The product focuses on turning financial data into:

```text
UNDERSTAND
    ↓
PREDICT
    ↓
SIMULATE
    ↓
EXPLAIN
    ↓
DECIDE

The core product principle is:

Finance Engine calculates. AI interprets. User decides.

# 2. Background

Personal finance applications commonly focus on recording financial activity.

Typical functionality includes:

income tracking,
expense tracking,
categories,
budgets,
charts,
and monthly reports.

These features help answer questions such as:

How much did I spend?
Where did my money go?

However, users often need answers to more important questions:

Why did my money run out faster this month?
Which spending behavior changed?
Will my balance remain healthy until the next income date?
Can I afford a large purchase without damaging my financial buffer?
What happens if my income drops next month?
How long can my current savings support my lifestyle?
What should I change if I want to reach a financial goal faster?

Traditional expense trackers generally show historical information.

Savio is designed to move beyond historical tracking toward financial understanding and decision support.

# 3. Problem Statement

Users may already know their:

income,
expenses,
account balances,
recurring bills,
and budgets.

But raw financial data alone does not necessarily help them make better decisions.

The main problems Savio addresses are:

## 3.1 Financial Data Without Context

A user may see:

Food        Rp2,400,000
Transport   Rp900,000
Shopping    Rp1,800,000

but still not know:

whether these numbers are unusual,
what caused the increase,
how they compare with historical patterns,
or what impact they have on future cashflow.
## 3.2 Historical Tracking Without Forward Visibility

Most personal finance tools explain what already happened.

Users also need visibility into what may happen next.

For example:

Current balance
Rp8,500,000

Upcoming rent
-Rp3,000,000

Recurring subscriptions
-Rp750,000

Expected salary
+Rp9,000,000

Estimated variable spending
-Rp4,200,000

Savio should turn these events into a projected cashflow timeline.

## 3.3 Financial Decisions Are Often Made Without Simulation

Many financial decisions involve uncertainty.

Examples:

buying an expensive device,
adding a new installment,
resigning from a job,
reducing working hours,
moving to a more expensive residence,
increasing monthly savings,
or taking on additional recurring expenses.

Users normally make these decisions mentally or with rough calculations.

Savio provides structured scenario simulation before the decision is made.

## 3.4 Generic Advice Is Often Not Useful

Generic advice such as:

Spend less.

Save more.

Avoid unnecessary expenses.

provides little value without context.

Useful financial guidance should instead be based on actual user data and explain:

what changed,
why it matters,
what contributes most,
and what alternative actions are available.
# 4. Product Thesis

Savio is based on the following thesis:

Financial tracking becomes significantly more useful when historical data is transformed into explainable insights, future cashflow projections, and decision simulations.

Instead of only asking:

What happened to my money?

Savio should also help users ask:

Why did it happen?

What happens next?

What happens if I change something?

# 5. Vision

Savio aims to become a personal financial intelligence layer that helps users understand and reason about their own cashflow.

The long-term vision is:

Make personal financial decisions more understandable, measurable, and intentional.

Savio should help users move from reactive financial behavior toward proactive financial planning.

# 6. Product Goals
## 6.1 Primary Goals

Savio should allow users to:

Record and organize financial transactions.
Understand income and expense patterns.
Detect meaningful changes in spending behavior.
Create and monitor budgets.
Track recurring financial obligations.
Define financial goals.
Forecast future cashflow.
Simulate hypothetical financial decisions.
Receive explainable AI-generated insights.
Ask natural-language questions about their financial data.
## 6.2 Engineering Goals

The system should demonstrate:

clear business workflows,
deterministic financial calculations,
secure authentication,
authorization,
relational data modeling,
background processing,
structured validation,
auditability,
concurrency handling where necessary,
reusable frontend architecture,
testable business logic,
and explainable AI integration.
# 7. Non-Goals

The initial Savio product is not intended to become:

a banking application,
a payment processor,
a cryptocurrency wallet,
a stock trading platform,
an investment trading advisor,
an automated tax filing system,
a credit scoring platform,
an accounting ERP,
a lending platform,
or an autonomous financial advisor.

Savio does not execute financial transactions on behalf of users.

The system provides:

information,
analysis,
simulation,
explanation,

but the user remains responsible for financial decisions.

# 8. Target Users
## 8.1 Primary User

An individual who:

receives one or more sources of income,
has recurring monthly expenses,
wants better visibility into personal cashflow,
wants to control discretionary spending,
has financial goals,
and occasionally needs to evaluate large financial decisions.
## 8.2 Example User Profiles
Salaried Professional

Needs to understand:

monthly spending,
recurring expenses,
savings rate,
and whether lifestyle expenses remain sustainable.
Freelancer

Has irregular income and therefore needs:

income variability analysis,
cash runway,
future balance projections,
and conservative scenario planning.
Young Professional

May be making decisions such as:

buying electronics,
moving residence,
taking installments,
building an emergency fund,
or saving for travel.
Household User

Future capability may allow multiple users to collaborate on shared finances.

Examples:

shared household expenses,
shared goals,
and shared budgets.
# 9. User Problems
9.1 "Where Did My Money Go?"

The user knows their balance decreased but does not understand the main drivers.

Savio identifies:

major categories,
unusual transactions,
spending increases,
and changes compared with previous periods.
9.2 "Why Is This Month More Expensive?"

Savio compares:

Current Period
vs
Historical Baseline

and highlights contributing factors.

Example:

Dining increased       +Rp850,000
Transport increased    +Rp310,000
One-time purchase      +Rp1,200,000
9.3 "Will I Have Enough Money Later?"

Savio projects future balance using:

current balances,
scheduled income,
recurring expenses,
expected transactions,
and estimated variable spending.
9.4 "Can I Afford This?"

The user can create a financial scenario.

Example:

Purchase:
Laptop Rp15,000,000

Savio compares:

Baseline
vs
With Purchase

and shows the impact on:

future balance,
cash runway,
savings progress,
and financial buffer.
9.5 "What Happens If My Income Changes?"

Users can simulate:

Income decreases by 30%

or:

Salary income stops after September

Savio recalculates the projected financial position.

9.6 "Can I Reach My Goal?"

Example:

Goal:
Rp30,000,000

Target:
10 months

Savio calculates:

required contribution,
current trajectory,
projected completion date,
and potential gap.
# 10. Value Proposition

Savio provides value through four layers.

## 10.1 Track

Capture financial activity.

Income
Expenses
Transfers
Recurring Transactions
## 10.2 Understand

Transform transaction data into financial context.

Spending patterns
Budget variance
Income composition
Cashflow trends
Financial health indicators
## 10.3 Predict

Estimate future financial conditions.

Projected balance
Upcoming cash pressure
Goal trajectory
Cash runway
## 10.4 Simulate

Evaluate hypothetical decisions.

Large purchase
Income loss
New recurring expense
Higher savings target
Lifestyle adjustment
# 11. Core Product Principles
## 11.1 Deterministic Financial Calculations

Financial calculations must be deterministic.

The application backend calculates:

balances,
income totals,
expense totals,
budget utilization,
savings rate,
financial ratios,
cashflow forecast baseline,
goal progress,
scenario impact,
and financial health metrics.

The LLM must not be responsible for producing authoritative financial numbers.

## 11.2 AI Explains, Not Invents

AI receives structured financial context generated by Savio.

Example:

{
  "monthly_income": 12000000,
  "monthly_expense": 8400000,
  "savings_rate": 0.30,
  "dining_change_percent": 42,
  "forecast_min_balance": 2100000
}

AI may explain:

what changed,
why it may matter,
what patterns are visible,
and what alternatives may exist.

AI must not independently invent financial values.

## 11.3 Explainability

Important insights should explain their basis.

Instead of:

Your finances are unhealthy.

Savio should show:

Your financial buffer decreased mainly because:

- Dining spending increased by 42%
- A one-time purchase added Rp1.2M
- Monthly savings fell from 24% to 11%
## 11.4 User Control

AI-generated content is advisory.

The user should be able to:

Accept,
Dismiss,
Review,
Edit,
Recalculate,

where relevant.

## 11.5 Financial Privacy

Financial data is sensitive.

Savio should follow strong security principles including:

secure authentication,
strict authorization,
minimal sensitive logging,
protected credentials,
secure cookie handling,
and explicit data ownership.
## 11.6 Progressive Intelligence

Savio should remain useful even without AI.

The deterministic platform must still provide:

transactions,
budgets,
goals,
analytics,
forecasting,
and scenario simulation.

AI enhances these capabilities rather than being required for basic system correctness.

# 12. Core Capabilities
## 12.1 Accounts

Users can maintain financial accounts such as:

Cash
Bank Account
E-Wallet
Savings
Other

Each account contains:

name,
type,
balance,
currency,
and status.
## 12.2 Transactions

Transaction types:

INCOME
EXPENSE
TRANSFER
ADJUSTMENT

Transactions contain:

account,
amount,
category,
transaction date,
description,
notes,
tags,
and optional recurring relation.
## 12.3 Categories

Categories organize transactions.

Examples:

Income
├── Salary
├── Freelance
├── Business
└── Other

Expense
├── Food
├── Transport
├── Housing
├── Utilities
├── Entertainment
├── Shopping
├── Healthcare
└── Other

Users may create custom categories.

## 12.4 Recurring Transactions

Users may define recurring:

income,
bills,
subscriptions,
installments,
savings contributions,
and other predictable cashflow.

Example:

Salary
Rp12,000,000
Monthly
Every 25th
## 12.5 Budgets

Users can define budgets:

Food
Rp1,500,000 / month

Savio tracks:

used
remaining
utilization
projected overspend
## 12.6 Financial Goals

Examples:

Emergency Fund
Rp30,000,000

Laptop
Rp20,000,000

Travel
Rp15,000,000

Goals include:

target amount,
target date,
current amount,
contribution strategy,
and progress.
## 12.7 Cashflow Analytics

Savio provides analytics including:

total income,
total expenses,
net cashflow,
savings rate,
category distribution,
spending trends,
income trends,
and period comparisons.
## 12.8 Cashflow Forecast

Forecasting uses deterministic financial information.

Inputs may include:

current account balances,
recurring income,
recurring expenses,
scheduled transactions,
average variable spending,
and user-defined assumptions.

Outputs include:

Projected daily / weekly balance
Lowest projected balance
Expected cashflow pressure
Forecast horizon
## 12.9 Scenario Simulator

Users can create hypothetical financial scenarios.

One-Time Expense
Buy laptop:
-Rp15,000,000
on September 15
Income Change
Salary:
-30%
starting October
Income Removal
Salary stops
starting November
Recurring Expense
New installment:
Rp1,500,000/month
for 12 months
Savings Adjustment
Increase saving:
+Rp1,000,000/month

Savio calculates the difference between:

BASELINE
vs
SCENARIO
## 12.10 AI Insights

AI may create structured financial insights based on computed data.

Insight types may include:

SPENDING_ANOMALY
INCOME_CHANGE
BUDGET_RISK
CASHFLOW_RISK
SAVINGS_PATTERN
GOAL_RISK
RECURRING_COST
POSITIVE_TREND

Example:

Dining spending is 42% higher than your 3-month average.

Primary contributor:
Food delivery +Rp720,000.
## 12.11 AI Financial Copilot

Users can ask questions in natural language.

Examples:

Why did I spend more this month?

What are my biggest recurring expenses?

How much do I usually spend after payday?

Can I afford a Rp10M purchase this month?

What happens if I lose my salary next month?

Which expenses changed the most?

The AI Copilot receives structured context from Savio rather than unrestricted raw database access.

## 12.12 Reports

Users can view summaries by:

week
month
quarter
year
custom range

Potential reports include:

income vs expense,
category breakdown,
savings rate,
budget performance,
goal progress,
and cashflow trend.
# 13. AI Collaboration Principles

AI in Savio is a collaborator rather than the financial authority.

The relationship is:

User Financial Data
        ↓
Finance Engine
        ↓
Analytics / Forecast / Scenario Engine
        ↓
AI Context Builder
        ↓
LLM
        ↓
Structured Insight
        ↓
User
## 13.1 AI Responsibilities

AI may:

categorize transaction descriptions,
summarize financial patterns,
explain unusual spending,
identify possible behavioral patterns,
explain forecast results,
compare scenarios,
propose budget adjustments,
translate analytics into natural language,
and answer questions from available financial context.
## 13.2 AI Must Not

AI must not:

modify balances directly,
create authoritative transactions without confirmation,
calculate authoritative balances,
silently change budgets,
execute payments,
provide guaranteed financial outcomes,
make investment trades,
or override deterministic financial rules.
## 13.3 Confidence-Based AI Actions

AI-assisted classification should expose confidence where appropriate.

Example:

Transaction:
"GRAB*FOOD 83219"

Suggested Category:
Food & Dining

Confidence:
96%

High-confidence recommendations may be easier to accept.

Low-confidence suggestions should require explicit user review.

## 13.4 Structured AI Output

AI responses used by application workflows should prefer structured output.

Example:

{
  "type": "spending_anomaly",
  "severity": "medium",
  "title": "Dining spending increased",
  "summary": "Dining spending is 42% above the recent baseline.",
  "drivers": [
    {
      "category": "Food Delivery",
      "difference": 720000
    }
  ],
  "suggested_actions": [
    {
      "type": "budget_review",
      "category": "Food & Dining"
    }
  ]
}
# 14. Killer Features

Savio is differentiated by three primary capabilities.

## 14.1 Explainable AI Insights

Savio should not simply show charts.

The system should explain:

What changed?
Why?
How significant is it?
What contributed most?
## 14.2 Cashflow Forecasting

Users should be able to see how their current financial pattern may affect future balances.

Example:

Today        Rp12.4M
Sep 01       Rp10.1M
Sep 10       Rp7.6M
Sep 20       Rp3.2M
Sep 25       Rp12.0M
## 14.3 Financial Scenario Simulator

Users can test a financial decision before applying it.

Example:

Scenario:
Buy Laptop Rp15M

Comparison:

                    BASELINE     SCENARIO

Month-end balance   Rp18.4M      Rp3.4M

Savings rate        27%          8%

Emergency buffer    4.1 months   1.9 months

Goal delay          0 months     3 months

AI then explains the implications.

# 15. Primary User Journey

A typical Savio journey is:

Register
   ↓
Create Financial Accounts
   ↓
Record / Import Transactions
   ↓
Configure Categories
   ↓
Define Recurring Transactions
   ↓
Create Budget
   ↓
Create Financial Goals
   ↓
View Cashflow Analytics
   ↓
Receive AI Insights
   ↓
View Cashflow Forecast
   ↓
Create Financial Scenario
   ↓
Compare Scenario
   ↓
Ask AI Copilot
   ↓
Make User Decision
# 16. Success Metrics

Initial product success should be evaluated using product behavior rather than vanity metrics.

## 16.1 Engagement

Potential metrics include:

percentage of users recording transactions regularly,
number of active budgets,
number of active financial goals,
forecast usage,
scenario simulations per user.
## 16.2 Insight Usefulness

Potential metrics include:

insight viewed,
insight dismissed,
insight accepted,
user feedback on insight relevance.
## 16.3 Financial Awareness

Potential product indicators include:

reduction in unexpected budget overruns,
improvement in goal tracking consistency,
increased awareness of recurring expenses,
frequency of scenario simulation before major expenses.

These metrics should not claim direct financial improvement unless supported by actual product data.

# 17. MVP Scope

The MVP should prove Savio's primary thesis without overbuilding.

## 17.1 P0 — Core

Required:

Authentication
User Profile

Accounts
Categories
Transactions

Income / Expense Tracking
Transfer Handling

Recurring Transactions

Budgets

Financial Goals

Cashflow Dashboard

Search
Filter
Sort
Pagination

Basic Cashflow Forecast

Scenario Simulator

AI Insight Generation

AI Copilot

Validation
Error Handling
Auditability
## 17.2 P1 — High Value

After core functionality is stable:

Advanced forecast assumptions
Spending anomaly detection
Budget overspend prediction
Goal feasibility analysis
Financial health indicators
AI transaction categorization
Insight feedback
Notifications
Background AI jobs
Reports
## 17.3 P2 — Enhancement

Future enhancements:

CSV transaction import
Receipt upload
AI receipt extraction
Household finance
Shared budgets
Shared goals
Multi-currency
Bank integration
Advanced recurring detection
Advanced anomaly detection
Export
Mobile application
# 18. Post-MVP Vision
## 18.1 Household Finance
Household
├── Owner
├── Member
└── Viewer

Shared:

expenses,
budgets,
goals,
and financial planning.
## 18.2 Transaction Import

Users may upload:

CSV
Excel
Bank statement

with data validation and review before commit.

## 18.3 Receipt Intelligence

Receipt upload:

Image
   ↓
AI / OCR extraction
   ↓
Merchant
Amount
Date
Items
Category
   ↓
User Review
   ↓
Transaction
## 18.4 Intelligent Recurring Detection

Savio may identify patterns such as:

Netflix
Rp186,000

appears every ~30 days

and suggest:

Possible recurring expense.
# 19. Risks & Constraints
## 19.1 AI Hallucination

AI may produce incorrect explanations.

Mitigation:

deterministic calculations,
structured context,
structured output,
confidence indicators,
human control,
and no autonomous financial execution.
## 19.2 Financial Advice Risk

Savio should avoid presenting AI output as guaranteed professional financial advice.

The product focuses on:

financial understanding
cashflow analysis
scenario comparison
decision support

rather than investment or legal financial advice.

## 19.3 Insufficient Historical Data

Forecast quality may be limited for new users.

Savio should expose when predictions are based on insufficient data.

Example:

Forecast confidence is limited because
only 18 days of transaction history are available.
## 19.4 Data Privacy

Financial data requires strong security controls.

Security design must be treated as a first-class architecture concern.

## 19.5 Forecast Uncertainty

Future transactions cannot always be predicted.

Savio must distinguish between:

KNOWN
SCHEDULED
ESTIMATED
ASSUMED

financial events.

# 20. Product Positioning

Savio should not be positioned as:

AI Expense Tracker

or:

Budget App with ChatGPT

The intended positioning is:

Savio is an AI-powered personal cashflow intelligence and financial decision support platform.

Short product statement:

Understand your money today. See what comes next. Test decisions before making them.

# 21. Product Pillars

Savio is built around four product pillars.

## 21.1 UNDERSTAND
Income
Expenses
Cashflow
Patterns
Budgets
Goals
## 21.2 PREDICT
Future balance
Cashflow pressure
Budget risk
Goal trajectory
## 21.3 SIMULATE
Large purchase
Income change
Recurring expense
Savings adjustment
Lifestyle decision
## 21.4 EXPLAIN
AI insights
Pattern explanation
Scenario explanation
Natural-language questions

Together:

UNDERSTAND
    ↓
PREDICT
    ↓
SIMULATE
    ↓
EXPLAIN
    ↓
USER DECIDES
# 22. Product Decision Rule

Whenever a new feature is proposed for Savio, evaluate it using the following questions:

Does this help the user understand their financial condition?
Does this improve future cashflow visibility?
Does this help simulate a meaningful financial decision?
Does AI provide real value here instead of being decorative?
Can the financial logic remain deterministic and testable?
Can the feature be explained clearly during technical review?
Does it improve Savio's core product thesis?

If the answer is consistently no, the feature probably does not belong in Savio.

# 23. Final Product Thesis

Savio is based on one central idea:

People do not only need a record of where their money went. They need help understanding why it happened, what is likely to happen next, and how today's decisions may change tomorrow's financial position.

Savio therefore combines:

financial tracking
+
deterministic financial intelligence
+
forecasting
+
scenario simulation
+
explainable AI

to create a personal financial decision support experience.

The final responsibility remains with the user:

Finance Engine calculates. AI interprets. User decides.


Nah, format ini **satu blok Markdown utuh**, jadi tinggal copy seluruh isi code block lalu paste langsung ke:

```text
docs/product/product-foundation.md