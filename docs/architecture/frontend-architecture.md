# Savio — Frontend Architecture

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Design System](../../DESIGN.md) — visual identity and UX guidelines for the UI.
- [System Architecture](system-architecture.md) — backend layers these pages consume.
- [API Contract](../api/api-contract.md) — endpoints, errors, and response mapping.
- [Business Requirements](../product/business-requirements.md) — business rules the UI must respect.

## 1. Document Purpose

This document defines the frontend architecture for Savio.

The purpose of this document is to translate Savio's:

- product foundation,
- business requirements,
- user flows,
- database design,
- API contract,
- AI architecture,
- and system architecture

into a concrete frontend implementation strategy.

This document defines:

- frontend technology stack,
- application structure,
- routing,
- information architecture,
- navigation,
- layout,
- design system,
- component architecture,
- feature organization,
- server-state management,
- form handling,
- API integration,
- authentication state,
- Axios interceptors,
- loading states,
- empty states,
- error states,
- responsive behavior,
- AI user experience,
- forecast visualization,
- scenario simulation UX,
- accessibility,
- testing,
- and frontend engineering conventions.

The core Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

The frontend displays and helps users interact with authoritative backend results.

It must not become a second implementation of Savio's financial calculation engine.

---

# 2. Frontend Technology Stack

Recommended frontend stack:

```text
React
TypeScript
Vite
Duralux SCSS (vendored Bootstrap UI theme)
React Router
Axios
TanStack Query
React Hook Form
Zod
```

Recommended testing stack:

```text
Vitest
React Testing Library
MSW
Playwright
```

Optional supporting libraries:

```text
react-icons
date-fns
Recharts
clsx
```

The exact chart library may change during implementation, but the frontend architecture should remain independent from a specific chart library.

---

# 3. Frontend Architecture Goals

The Savio frontend should be:

```text
modular
predictable
responsive
accessible
type-safe
testable
consistent
```

The architecture must support both:

```text
simple financial data entry
```

and:

```text
complex financial intelligence workflows
```

such as:

- cashflow forecast,
- budget analysis,
- scenario comparison,
- AI insights,
- AI Copilot.

---

# 4. Frontend Responsibility Boundary

The frontend owns:

```text
presentation

navigation

interaction

client-side validation

form state

server-state synchronization

loading states

empty states

error states

formatting

responsive UX

accessibility
```

The frontend does not own authoritative:

```text
account balances

budget utilization

savings rate

goal feasibility

cash runway

forecast calculations

scenario calculations

financial health scores
```

These values come from the backend.

---

# 5. Core Frontend Principle

The frontend should never independently reconstruct financial truth from raw data when the backend already provides an authoritative domain result.

Bad:

```text
Frontend downloads all transactions
↓
Frontend calculates savings rate
↓
Frontend decides budget status
```

Preferred:

```text
Frontend requests analytics
↓
Backend Finance Engine calculates
↓
Frontend displays result
```

---

# 6. High-Level Frontend Architecture

```text
Browser
   ↓
React Application
   │
   ├── Router
   │
   ├── Application Shell
   │
   ├── Feature Modules
   │
   ├── Shared UI
   │
   ├── TanStack Query
   │
   ├── React Hook Form
   │
   └── Axios Client
   │
   ▼
Savio REST API
```

---

# 7. Application Layers

Recommended conceptual frontend layers:

```text
App Layer
│
├── routing
├── providers
├── layouts
└── bootstrap

Feature Layer
│
├── auth
├── dashboard
├── accounts
├── transactions
├── recurring
├── budgets
├── goals
├── analytics
├── forecast
├── scenarios
├── insights
├── copilot
├── notifications
└── settings

Shared Layer
│
├── UI components
├── API client
├── utilities
├── hooks
├── formatting
├── types
└── validation helpers
```

---

# 8. Recommended Frontend Directory Structure

```text
frontend/
├── public/
│
├── src/
│   ├── app/
│   │   ├── router/
│   │   │   ├── index.tsx
│   │   │   ├── routes.tsx
│   │   │   ├── protected-route.tsx
│   │   │   └── guest-route.tsx
│   │   │
│   │   ├── providers/
│   │   │   ├── app-providers.tsx
│   │   │   ├── query-provider.tsx
│   │   │   └── auth-provider.tsx
│   │   │
│   │   ├── layouts/
│   │   │   ├── app-layout.tsx
│   │   │   ├── auth-layout.tsx
│   │   │   └── onboarding-layout.tsx
│   │   │
│   │   ├── config/
│   │   │   ├── navigation.ts
│   │   │   └── env.ts
│   │   │
│   │   └── main.tsx
│   │
│   ├── features/
│   │   ├── auth/
│   │   ├── onboarding/
│   │   ├── dashboard/
│   │   ├── accounts/
│   │   ├── categories/
│   │   ├── transactions/
│   │   ├── recurring/
│   │   ├── budgets/
│   │   ├── goals/
│   │   ├── analytics/
│   │   ├── forecast/
│   │   ├── scenarios/
│   │   ├── insights/
│   │   ├── copilot/
│   │   ├── notifications/
│   │   └── settings/
│   │
│   ├── shared/
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   ├── errors.ts
│   │   │   └── types.ts
│   │   │
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   ├── feedback/
│   │   │   ├── forms/
│   │   │   ├── data-display/
│   │   │   └── layout/
│   │   │
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── utils/
│   │   ├── constants/
│   │   ├── types/
│   │   └── validation/
│   │
│   ├── assets/
│   ├── styles/
│   └── vite-env.d.ts
│
├── tests/
│   ├── setup.ts
│   ├── mocks/
│   └── e2e/
│
├── index.html
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.ts
└── package.json
```

---

# 9. Feature Module Structure

Each large feature should contain only the pieces it owns.

Example:

```text
features/transactions/
├── api/
│   ├── transaction.api.ts
│   └── transaction.query-keys.ts
│
├── components/
│   ├── transaction-form.tsx
│   ├── transaction-table.tsx
│   ├── transaction-filters.tsx
│   ├── transaction-type-badge.tsx
│   └── transaction-detail-panel.tsx
│
├── hooks/
│   ├── use-transactions.ts
│   ├── use-create-transaction.ts
│   ├── use-update-transaction.ts
│   └── use-void-transaction.ts
│
├── pages/
│   ├── transaction-list-page.tsx
│   └── transaction-detail-page.tsx
│
├── schemas/
│   └── transaction.schema.ts
│
├── types/
│   └── transaction.types.ts
│
└── utils/
    └── transaction-formatters.ts
```

---

# 10. Feature Boundary Rule

A feature owns:

```text
feature-specific API

feature-specific UI

feature-specific hooks

feature-specific validation

feature-specific types
```

Shared concepts belong in:

```text
shared/
```

Avoid moving everything into `shared` too early.

---

# 11. Route Architecture

Primary application route hierarchy:

```text
/
├── login
├── register
│
├── onboarding
│
└── app
    ├── dashboard
    ├── accounts
    │   └── :accountId
    │
    ├── transactions
    │   └── :transactionId
    │
    ├── recurring
    │   └── :recurringId
    │
    ├── budgets
    │   └── :budgetId
    │
    ├── goals
    │   └── :goalId
    │
    ├── analytics
    ├── forecast
    ├── scenarios
    │   └── :scenarioId
    │
    ├── insights
    │   └── :insightId
    │
    ├── copilot
    ├── notifications
    └── settings
```

---

# 12. Recommended Final URL Structure

Use concise user-facing routes:

```text
/login

/register

/onboarding

/dashboard

/accounts
/accounts/:accountId

/transactions
/transactions/:transactionId

/recurring
/recurring/:recurringId

/budgets
/budgets/:budgetId

/goals
/goals/:goalId

/analytics

/forecast

/scenarios
/scenarios/:scenarioId

/insights
/insights/:insightId

/copilot

/notifications

/settings/profile
/settings/preferences
/settings/security
/settings/sessions
```

---

# 13. Root Route Behavior

Route:

```text
/
```

Behavior:

```text
Authenticated?
├── Yes → /dashboard
└── No  → /login
```

---

# 14. Guest Routes

Guest-only routes:

```text
/login

/register
```

If authenticated user opens `/login`:

```text
redirect /dashboard
```

---

# 15. Protected Routes

Protected:

```text
/dashboard

/accounts

/transactions

/recurring

/budgets

/goals

/analytics

/forecast

/scenarios

/insights

/copilot

/notifications

/settings
```

Unauthenticated access:

```text
redirect /login
```

---

# 16. Onboarding Guard

Authenticated users who have not completed minimum onboarding may be redirected to:

```text
/onboarding
```

Example condition:

```text
no financial account exists
```

The exact onboarding completion state should preferably come from backend profile/dashboard bootstrap rather than being inferred only client-side.

---

# 17. Main Navigation

Recommended main navigation:

```text
Overview
├── Dashboard
└── Analytics

Money
├── Accounts
├── Transactions
└── Recurring

Planning
├── Budgets
├── Goals
├── Forecast
└── Scenarios

Intelligence
├── AI Insights
└── Savio Copilot

System
└── Settings
```

---

# 18. Sidebar Proposal

Desktop sidebar:

```text
SAVIO

Overview
  Dashboard
  Analytics

Money
  Accounts
  Transactions
  Recurring

Planning
  Budgets
  Goals
  Forecast
  Scenarios

Intelligence
  AI Insights
  Savio Copilot

Settings
```

Notifications may live primarily in the header rather than sidebar.

---

# 19. Why This Information Architecture

The grouping follows the user mental model:

```text
What happened?
→ Overview / Money

What am I planning?
→ Planning

What does Savio understand?
→ Intelligence
```

This is preferable to a flat list of 12 unrelated menu items.

---

# 20. Application Shell

Desktop layout:

```text
┌────────────────────────────────────────────────────────────┐
│ Sidebar │ Header                                           │
│         ├──────────────────────────────────────────────────┤
│         │                                                  │
│         │                 Page Content                     │
│         │                                                  │
│         │                                                  │
└─────────┴──────────────────────────────────────────────────┘
```

---

# 21. Sidebar Responsibilities

Sidebar contains:

```text
Savio brand

Primary navigation

Grouped sections

Active route state

Collapse behavior

User/settings shortcut
```

---

# 22. Header Responsibilities

Header may contain:

```text
Page title / breadcrumb

Quick Add button

Notification button

AI Copilot shortcut

User menu
```

---

# 23. Global Quick Add

A high-value global action:

```text
+ Add
```

Menu:

```text
Income

Expense

Transfer

Recurring Transaction

Budget

Goal
```

On mobile, this may become:

```text
floating / prominent add action
```

---

# 24. Quick Transaction Entry

Because financial tracking is frequent, Savio should reduce friction.

Desktop:

```text
Quick Add
→ Add Expense
```

may open:

```text
modal / sheet
```

rather than requiring full page navigation.

More complex edits can use dedicated routes or detail sheets.

---

# 25. Desktop Layout

Recommended:

```text
Sidebar:
240–280 px

Header:
64–72 px

Content:
responsive centered/max width where appropriate
```

Data-heavy pages such as transactions may use wider layouts.

---

# 26. Mobile Navigation

On mobile, desktop sidebar should not simply shrink.

Recommended:

```text
Bottom Navigation
+
More Drawer
```

Potential bottom navigation:

```text
Home
Transactions
Forecast
Copilot
More
```

---

# 27. Mobile More Menu

`More` contains:

```text
Accounts
Recurring
Budgets
Goals
Scenarios
Insights
Analytics
Settings
```

Exact navigation may be adjusted after visual prototyping.

---

# 28. Responsive Breakpoints

Tailwind defaults may be used:

```text
sm

md

lg

xl

2xl
```

Design should be mobile-aware rather than desktop-only with emergency wrapping.

---

# 29. Page Layout Pattern

Most pages should follow:

```text
Page Header
├── Title
├── Description
└── Actions

Optional Filters

Main Content

Pagination / Secondary Content
```

---

# 30. Page Header Component

Shared component concept:

```tsx
<PageHeader
  title="Transactions"
  description="Track and review your financial activity."
  actions={<AddTransactionButton />}
/>
```

---

# 31. Dashboard Information Architecture

Dashboard should answer the most important questions first.

Recommended order:

```text
1. Financial Position
2. Cashflow
3. Forecast
4. Budget Risks
5. AI Insights
6. Goals
7. Upcoming Activity
```

---

# 32. Dashboard Desktop Layout

Concept:

```text
┌─────────────────────────────────────────────────────────────┐
│ Total Balance   Income   Expense   Net Cashflow             │
├──────────────────────────────────────┬──────────────────────┤
│ Cashflow Trend                       │ Forecast Snapshot    │
│                                      │                      │
├──────────────────────────────────────┼──────────────────────┤
│ Budget Status                        │ Goals                │
├──────────────────────────────────────┴──────────────────────┤
│ Savio Insights                                              │
├─────────────────────────────────────────────────────────────┤
│ Upcoming Financial Events                                  │
└─────────────────────────────────────────────────────────────┘
```

---

# 33. Dashboard Summary Cards

Cards may include:

```text
Total Balance

Income This Month

Expenses This Month

Net Cashflow
```

Secondary metric:

```text
Savings Rate
```

Avoid overloading the first row with too many numbers.

---

# 34. Dashboard Financial Value Formatting

Example:

```text
Rp16.250.000
```

Secondary abbreviated form may be used selectively:

```text
Rp16,25 jt
```

but full values should remain accessible where precision matters.

---

# 35. Money Formatting Utility

Create a centralized utility:

```text
formatMoney()
```

Input:

```text
"16250000.00"
IDR
```

Output:

```text
Rp16.250.000
```

Do not manually format money throughout components.

---

# 36. Percentage Formatting

Central utility:

```text
formatPercent()
```

Examples:

```text
30.00
→ 30%

23.53
→ 23,53%
```

respecting locale.

---

# 37. Date Formatting

Use centralized date formatting.

Examples:

```text
24 Aug 2026

24 August 2026
```

depending on context.

Never repeatedly implement date formatting inline.

---

# 38. Account Page

Accounts list should show:

```text
Account Name

Type

Current Balance

Currency

Status
```

Possible view:

```text
cards
```

rather than a dense table because account count is normally small.

---

# 39. Account Card

Example:

```text
BCA Main

Bank

Rp14.750.000

Active
```

Actions:

```text
View

Edit

Reconcile

Archive
```

---

# 40. Account Detail Page

Sections:

```text
Account Overview

Balance Summary

Income / Expense Summary

Balance Trend

Recent Transactions

Transfer Activity
```

Actions:

```text
Add Transaction

Transfer

Reconcile

Edit

Archive
```

---

# 41. Account Reconciliation UX

Flow:

```text
Reconcile Balance
```

Modal:

```text
Tracked Balance
Rp4.800.000

Actual Balance
[ Rp5.000.000 ]

Difference
+Rp200.000

Reason
[ ... ]
```

Before confirmation, explain:

```text
Savio will create an adjustment record rather than overwrite financial history.
```

---

# 42. Transactions Page

Transactions is one of the most frequently used pages.

Desktop layout:

```text
Page Header
   ↓
Summary / Quick Filters
   ↓
Search + Filters
   ↓
Transaction Table
   ↓
Pagination
```

---

# 43. Transaction Table Columns

Recommended:

```text
Date

Description

Account

Category

Type

Amount

Status

Actions
```

Avoid unnecessary columns.

---

# 44. Transaction Amount Styling

Conceptually:

```text
Income
+Rp12.000.000

Expense
-Rp87.500
```

Do not rely on color alone to indicate direction.

Use:

```text
sign
+
text label / icon where useful
```

for accessibility.

---

# 45. Transaction Filters

Filter controls:

```text
Search

Type

Account

Category

Date Range

Amount Range

Status
```

Sort:

```text
Newest

Oldest

Highest Amount

Lowest Amount
```

---

# 46. Transaction Filter UX

Desktop:

```text
search bar
+
compact filter controls
```

Mobile:

```text
Search
[ Filter ]
```

Filter button opens bottom sheet/drawer.

---

# 47. URL-Synchronized Filters

Important filters should be represented in URL query params.

Example:

```text
/transactions?type=EXPENSE&category=food&page=2
```

Benefits:

```text
refresh persistence

browser back support

shareability

predictable state
```

---

# 48. Transaction Creation UI

Recommended transaction form fields:

```text
Type

Account

Amount

Category

Date

Description

Merchant

Notes
```

The most important fields should appear first.

---

# 49. Transaction Type Selector

Use explicit segmented control:

```text
Income | Expense
```

Transfer should ideally use a separate flow because it requires source and destination accounts.

Adjustment should normally be accessible through account reconciliation, not ordinary transaction creation.

---

# 50. Amount Input

Amount field should support:

```text
localized visual formatting
```

while storing a clean decimal representation.

Avoid direct JavaScript floating-point arithmetic.

---

# 51. AI Category Suggestion UI

When description/merchant is entered:

```text
Suggested by Savio AI

Food & Dining
High confidence

[Use Category]
```

The UI should clearly communicate:

```text
suggestion
```

not:

```text
final classification
```

---

# 52. AI Category Failure UX

If suggestion fails:

```text
AI suggestion unavailable.
Choose a category manually.
```

Do not block transaction creation.

---

# 53. Transaction Detail

Detail view may use:

```text
side sheet
```

on desktop for fast review, or dedicated route when deep linking is useful.

Content:

```text
amount

type

date

account

category

merchant

description

notes

source

status

created time
```

Actions:

```text
Edit

Void
```

---

# 54. Void Transaction UX

Confirmation:

```text
Void this transaction?

The original financial effect will be reversed and your account balance will be updated.

Reason:
[                         ]

[Cancel] [Void Transaction]
```

This should feel more intentional than a generic delete.

---

# 55. Transfer UI

Separate form:

```text
From Account

To Account

Amount

Date

Description
```

Display before confirmation:

```text
BCA Main
Rp5.000.000
        ↓ Rp1.000.000
GoPay
Rp300.000
```

Projected after:

```text
BCA Main
Rp4.000.000

GoPay
Rp1.300.000
```

The preview may be visual, but backend remains authoritative.

---

# 56. Recurring Page

Recurring page groups:

```text
Income

Bills

Subscriptions

Other Recurring
```

or filter by:

```text
ALL

INCOME

EXPENSE
```

---

# 57. Recurring Card

Example:

```text
Salary

Rp12.000.000

Monthly · Every 25th

Next:
25 Sep 2026

Auto Post

ACTIVE
```

Actions:

```text
Edit

Pause

End
```

---

# 58. Recurring Status UX

Statuses:

```text
ACTIVE

PAUSED

ENDED
```

Use consistent badges.

Do not make ended rules look editable as active rules.

---

# 59. Recurring Occurrence History

Detail page may show:

```text
25 Aug 2026
POSTED

25 Jul 2026
POSTED
```

Useful for demonstrating idempotent recurring processing.

---

# 60. Budgets Page

Budget overview should prioritize progress.

Example card:

```text
Food & Dining

Rp1.450.000 / Rp2.000.000

72.5%

Rp550.000 remaining

ON TRACK
```

---

# 61. Budget Progress Bar

Progress visually communicates utilization.

States:

```text
ON_TRACK

WARNING

EXCEEDED
```

Do not rely on color alone.

Always include text.

---

# 62. Projected Budget Risk

Budget card/detail may show:

```text
Projected Month-End Spend
Rp2.350.000

Projected Overspend
Rp350.000
```

Distinguish clearly:

```text
Actual
```

from:

```text
Projected
```

---

# 63. Budget Detail

Sections:

```text
Budget Summary

Actual Spending

Projected Spending

Historical Comparison

Transactions in Category

AI Explanation if available
```

---

# 64. Goals Page

Goals may use cards because active goal count is typically small.

Example:

```text
Emergency Fund

Rp12M / Rp30M

40%

Target:
Feb 2027

Required:
Rp3M/month

AT RISK
```

---

# 65. Goal Detail

Sections:

```text
Progress

Target

Required Contribution

Estimated Free Cashflow

Feasibility

Historical Progress

Related Scenario Impact
```

---

# 66. Goal Feasibility UX

Never show only:

```text
AT RISK
```

Explain deterministic inputs:

```text
Required Contribution
Rp3M/month

Estimated Free Cashflow
Rp2.1M/month

Gap
Rp900k/month
```

Optional AI explanation can follow.

---

# 67. Analytics Page

Analytics should not become a wall of charts.

Recommended sections:

```text
Period Selector

Income vs Expense

Net Cashflow

Savings Rate

Expense Category Breakdown

Period Comparison

Largest Changes
```

---

# 68. Analytics Period Selector

Options:

```text
This Month

Last Month

3 Months

6 Months

This Year

Custom
```

Dates should be synchronized to query parameters where reasonable.

---

# 69. Analytics Chart Principles

Charts should:

```text
answer a question
```

not exist only for decoration.

Examples:

```text
How has cashflow changed?

Which categories dominate expenses?

What changed vs previous period?
```

---

# 70. Chart Accessibility

Provide:

```text
labels

tooltips

text summaries

table fallback where useful
```

Never make key financial meaning available only through chart geometry.

---

# 71. Forecast Page

Forecast is one of Savio's primary differentiated features.

The page should make future cashflow understandable without pretending certainty.

---

# 72. Forecast Page Structure

Recommended:

```text
Page Header

Forecast Controls

Summary Metrics

Projected Balance Chart

Financial Event Timeline

Assumptions

Confidence / Data Quality
```

---

# 73. Forecast Controls

Controls:

```text
30 Days

60 Days

90 Days

6 Months

12 Months
```

Optional:

```text
Assumptions
```

---

# 74. Forecast Summary

Display:

```text
Current Balance

Projected Ending Balance

Lowest Projected Balance

Projected Income

Projected Expense

Forecast Confidence
```

---

# 75. Forecast Visual Distinction

Forecast UI must distinguish:

```text
KNOWN

SCHEDULED

ESTIMATED

ASSUMED
```

Example legend:

```text
Confirmed

Recurring

Estimated

Assumption
```

Use user-friendly labels while preserving domain meaning.

---

# 76. Forecast Confidence UX

Example:

```text
Forecast confidence: Medium
```

Explanation:

```text
Based on 3 months of transaction history.
Approximately 62% of projected events are recurring or known.
```

Avoid pretending the confidence value is scientifically precise if the model is rule-based.

---

# 77. Forecast Chart

Primary chart:

```text
Projected Balance Over Time
```

Use:

```text
line chart
```

with markers for important events.

Events:

```text
Salary

Rent

Installment

Potential low-balance date
```

---

# 78. Forecast Timeline

Below chart:

```text
01 Sep
Rent
-Rp3M
Scheduled

10 Sep
Internet
-Rp450k
Scheduled

Estimated Daily Spending
-Rp...
Estimated
```

This makes forecast explainable.

---

# 79. Forecast Low-Balance State

If low balance is detected:

```text
Potential cashflow pressure
```

Card:

```text
Lowest projected balance:
Rp850.000

Expected:
18 Sep

Next salary:
25 Sep
```

Action:

```text
Explore Scenario
```

---

# 80. Forecast Empty State

Insufficient data:

```text
Savio needs more financial activity to build a useful forecast.

Add recurring income and expenses to improve the projection.
```

Actions:

```text
Add Recurring Income

Add Recurring Expense
```

---

# 81. Scenario Simulator

Scenario Simulator is the primary killer feature.

It should not feel like a generic CRUD form.

It should feel like an interactive financial decision workspace.

---

# 82. Scenario Page Layout

Desktop concept:

```text
┌──────────────────────────────┬──────────────────────────────┐
│ Scenario Builder             │ Scenario Comparison          │
│                              │                              │
│ Name                         │ Baseline                     │
│ Horizon                      │ vs                           │
│ Modifications                │ Scenario                     │
│                              │                              │
│ + Add Change                 │ Metrics                      │
│                              │ Chart                        │
│ [Calculate]                  │ Goal Impact                  │
└──────────────────────────────┴──────────────────────────────┘

AI Explanation
```

---

# 83. Scenario Mobile Layout

Mobile:

```text
Scenario Builder
   ↓
Calculate
   ↓
Comparison Summary
   ↓
Detailed Comparison
   ↓
AI Explanation
```

Do not force two-column desktop structure onto mobile.

---

# 84. Scenario Creation

Step 1:

```text
Scenario Name

Forecast Horizon
```

Example:

```text
Buy Laptop

6 Months
```

Then:

```text
+ Add Change
```

---

# 85. Scenario Modification Selector

Options:

```text
One-Time Expense

One-Time Income

Recurring Expense

Recurring Income

Reduce Income

Remove Income

Reduce Expense

Change Savings
```

Use user-friendly labels while mapping to backend enum values.

---

# 86. One-Time Purchase Form

Fields:

```text
Name

Amount

Date
```

Example:

```text
MacBook Pro

Rp15.000.000

10 Sep 2026
```

---

# 87. Income Reduction Form

Fields:

```text
Income Source

Reduction Percentage

Effective Date
```

Example:

```text
Salary

30%

1 Oct 2026
```

---

# 88. Income Removal Form

Example:

```text
Stop Salary

Starting:
1 Oct 2026
```

Useful shortcut label:

```text
"What if I resign?"
```

---

# 89. Recurring Expense Scenario Form

Fields:

```text
Name

Amount

Frequency

Start Date

Duration
```

Example:

```text
Motorcycle Installment

Rp1.800.000

Monthly

24 months
```

---

# 90. Scenario Calculation State

While calculating:

```text
Calculating scenario...
```

The UI may show steps:

```text
Building baseline

Applying changes

Comparing outcomes
```

But avoid fake progress percentages unless real.

---

# 91. Scenario Comparison

Core comparison:

```text
Metric                 Baseline       Scenario

Ending Balance         Rp18.4M        Rp3.4M

Lowest Balance         Rp8.2M         Rp1.1M

Savings Rate           27%            8%

Cash Runway            4.1 months     1.9 months

Goal Completion        Dec 2026       Mar 2027
```

---

# 92. Scenario Difference Emphasis

Also show:

```text
Difference

Ending Balance
-Rp15M

Cash Runway
-2.2 months

Goal Delay
+3 months
```

This makes consequences immediately understandable.

---

# 93. Scenario Chart

Overlay:

```text
Baseline projection

Scenario projection
```

on one chart.

The chart should make divergence visible.

---

# 94. Scenario Assumptions

Always show:

```text
Based on:
- current account balances
- recurring income and expenses
- recent variable spending
- scenario changes
```

This improves trust.

---

# 95. Scenario Stale UX

If source data changes:

```text
This scenario is out of date.

Your financial records changed after the last calculation.

[Recalculate]
```

Do not silently present stale results as current.

---

# 96. AI Scenario Explanation

After deterministic calculation:

```text
Savio AI Explanation
```

Example:

```text
This purchase does not immediately create a negative balance, but it reduces your lowest projected balance from Rp8.2M to Rp1.1M and delays your emergency fund target by approximately three months.
```

Display AI content below deterministic comparison, not above it.

This visually reinforces:

```text
facts first
AI interpretation second
```

---

# 97. Scenario Alternative Exploration

A high-value UX pattern:

```text
Compare Alternative
```

Examples:

```text
Buy today

Buy next month

Buy after saving Rp5M more
```

Each alternative remains a deterministic scenario.

AI may suggest alternatives, but user chooses which to calculate.

---

# 98. AI Insights Page

AI Insights should not look like chat.

It is a structured intelligence feed.

Filters:

```text
All

Cashflow

Budget

Spending

Goals

Positive Trends
```

---

# 99. AI Insight Card

Example:

```text
MEDIUM

Dining spending increased

Your dining spending is 60% above the recent baseline.

Main driver:
Food delivery +Rp720k

[View Details]
```

---

# 100. AI Insight Detail

Structure:

```text
Title

Severity

Summary

What Changed

Supporting Facts

Drivers

Suggested Actions

Generated At

Feedback
```

---

# 101. Insight Facts First

Example:

```text
Current:
Rp2.4M

3-Month Average:
Rp1.5M

Change:
+60%
```

Then:

```text
Savio AI explains:
...
```

Again:

```text
deterministic facts
before
AI narrative
```

---

# 102. AI Insight Actions

Allowed UI actions:

```text
Review Budget

View Transactions

Open Forecast

Create Scenario

Review Goal
```

These are generated from allowlisted backend action types.

---

# 103. Insight Feedback

Actions:

```text
Helpful

Not Helpful

Dismiss
```

Feedback should remain lightweight.

---

# 104. Savio Copilot

Copilot is a secondary interface to Savio.

It should not replace the primary structured application.

Use it for:

```text
questions

explanations

scenario entry

financial exploration
```

---

# 105. Copilot Route

```text
/copilot
```

Potential UI:

```text
Conversation Area

Suggested Questions

Message Composer

Optional Context / Result Cards
```

---

# 106. Copilot Welcome State

Suggested prompts:

```text
Why did I spend more this month?

What are my biggest recurring expenses?

Which budget is most at risk?

What does my cashflow look like next month?

Can I simulate buying a Rp10M laptop?

Am I on track for my emergency fund?
```

---

# 107. Copilot Response Design

Do not display only a text bubble when structured data exists.

Example:

```text
Savio

Your spending increased by Rp1.6M compared with your recent baseline.

Key drivers:

Food & Dining
+Rp700k

Shopping
+Rp550k

Transport
+Rp250k

[View Transactions]
```

---

# 108. Copilot Fact Cards

Backend `facts` can render as cards.

This prevents financial numbers from being buried in prose.

---

# 109. Copilot Clarification UX

If scenario date missing:

```text
When should I assume the purchase happens?

[Today]

[Next Payday]

[Choose Date]
```

Prefer structured choice controls instead of making the user type everything.

---

# 110. Copilot Scenario Handoff

If question becomes scenario:

```text
Can I afford a Rp15M laptop?
```

Copilot may:

```text
collect missing inputs
↓
calculate temporary scenario
↓
show comparison
↓
offer:
Save as Scenario
```

The deterministic scenario engine remains the source.

---

# 111. Copilot AI Failure State

Example:

```text
Savio AI is temporarily unavailable.

Your accounts, transactions, forecast, and scenarios are still available.
```

Actions:

```text
Try Again

Open Forecast

Open Scenarios
```

---

# 112. Notifications UX

Header notification icon:

```text
bell + unread badge
```

Popover may show latest notifications.

Dedicated page:

```text
/notifications
```

for full history.

---

# 113. Notification Item

Example:

```text
Food budget is almost reached

You have used 82% of your monthly Food & Dining budget.

2h ago
```

Click opens relevant resource.

---

# 114. Settings IA

Recommended settings routes:

```text
/settings/profile

/settings/preferences

/settings/security

/settings/sessions
```

Potential future:

```text
/settings/notifications

/settings/ai
```

AI preferences may initially live under general preferences.

---

# 115. Profile Settings

Fields:

```text
Name

Email display

Timezone

Locale
```

Email update may require separate secure workflow if implemented.

---

# 116. Financial Preferences

Fields:

```text
Default Currency

Budget Warning Threshold

Low Balance Threshold
```

---

# 117. AI Preferences

Toggles:

```text
AI Insights

Savio Copilot
```

Explain:

```text
Savio's core financial tools remain available even when AI features are disabled.
```

---

# 118. Security Settings

Potential:

```text
Password Change

Active Sessions

Logout All Sessions
```

---

# 119. Session List

Example:

```text
MacBook · Chrome

Current device

Last active:
Now


iPhone · Safari

Last active:
2 hours ago

[Revoke]
```

---

# 120. Authentication Bootstrap

Application startup must determine auth state safely.

State model:

```text
UNKNOWN

AUTHENTICATED

UNAUTHENTICATED
```

Do not initially assume unauthenticated before `/auth/me` resolves, or protected routes may flicker.

---

# 121. Auth Provider

Recommended conceptual API:

```tsx
const {
  user,
  status,
  logout
} = useAuth();
```

Where:

```text
status =
loading
authenticated
unauthenticated
```

---

# 122. Auth Source of Truth

Current authenticated user comes from:

```text
GET /api/v1/auth/me
```

Do not persist user authentication truth in localStorage.

---

# 123. Allowed Browser Storage

Authentication tokens:

```text
DO NOT STORE
```

in:

```text
localStorage
sessionStorage
```

Non-sensitive presentation preferences may be stored if useful.

Example:

```text
sidebar collapsed
theme
```

but not security state.

---

# 124. Axios Client

Central instance:

```text
shared/api/client.ts
```

Configuration:

```ts
axios.create({
  baseURL: API_URL,
  withCredentials: true
})
```

---

# 125. Request Interceptor

Potential responsibilities:

```text
attach CSRF header for mutations

attach request metadata if needed
```

No Authorization bearer token from browser storage is required if access token is cookie-based.

---

# 126. CSRF Token Access

Frontend may read:

```text
csrf_token cookie
```

or use a dedicated CSRF bootstrap state.

Mutation:

```text
X-CSRF-Token
```

must be attached.

---

# 127. Safe Mutation Methods

Apply CSRF to:

```text
POST

PUT

PATCH

DELETE
```

---

# 128. Axios 401 Architecture

Single-flight refresh is mandatory.

Pseudo flow:

```text
Request A → 401
Request B → 401
Request C → 401

          ↓

Is refresh running?
    ├── No → create refreshPromise
    └── Yes → wait refreshPromise

          ↓

ONE /auth/refresh

          ↓

Success:
retry A/B/C once

Failure:
reject all
clear auth state
redirect login
```

---

# 129. Refresh Promise

Conceptual module-level variable:

```ts
let refreshPromise: Promise<void> | null = null
```

Exact implementation may vary.

---

# 130. Retry Marker

Original request config must track:

```text
already retried?
```

Example custom property:

```ts
_retry
```

Never retry indefinitely.

---

# 131. Refresh Endpoint Exclusion

The refresh request itself must not trigger refresh interceptor recursively.

Likewise:

```text
login
register
```

must be handled carefully.

---

# 132. 403 Handling

On:

```text
403
```

do not automatically logout.

Possible reasons:

```text
permission denied

CSRF invalid
```

Use API error code to determine UI behavior.

---

# 133. 422 Handling

Do not convert all `422` errors into global toast.

Forms need access to:

```text
error.details
```

for field mapping.

---

# 134. 429 Handling

For rate limit:

```text
show clear feedback
```

If:

```text
Retry-After
```

exists, UI may show countdown for AI actions.

---

# 135. 500 Handling

Show generic safe error.

Example:

```text
Something went wrong.

Reference:
req_abc123
```

Never render backend trace.

---

# 136. API Error Type

Shared type:

```ts
interface ApiError {
  success: false
  error: {
    code: string
    details?: Record<string, string[]>
  }
  message: string
}
```

---

# 137. Server State

Use TanStack Query for:

```text
accounts

transactions

budgets

goals

analytics

forecast

scenarios

insights

notifications

settings
```

---

# 138. Query Key Factory

Avoid stringly-typed scattered keys.

Example:

```ts
export const transactionKeys = {
  all: ['transactions'] as const,
  lists: () => [...transactionKeys.all, 'list'] as const,
  list: (filters: TransactionFilters) =>
    [...transactionKeys.lists(), filters] as const,
  detail: (id: string) =>
    [...transactionKeys.all, 'detail', id] as const,
}
```

---

# 139. Mutation Strategy

Example:

```text
create transaction
```

On success:

```text
invalidate transaction list

invalidate account summary

invalidate dashboard

invalidate analytics

invalidate budgets

invalidate forecast freshness

invalidate scenario state where necessary
```

---

# 140. Avoid Manual Global State Sync

Do not manually update many separate global stores after every financial mutation unless necessary.

TanStack Query invalidation/refetch is safer for initial implementation.

---

# 141. Optimistic UI

Use optimistic updates selectively.

Safe candidates:

```text
notification mark read

insight dismiss
```

Be cautious for:

```text
account balances

transactions

transfers

scenario results
```

because financial correctness is more important than perceived instant updates.

---

# 142. Financial Mutation UX

For critical financial mutation:

```text
submit
↓
pending
↓
backend confirms
↓
invalidate/refetch
↓
display authoritative state
```

This is acceptable.

---

# 143. Form Architecture

Use:

```text
React Hook Form
+
Zod
```

Each feature owns its form schema.

Example:

```text
features/transactions/schemas/transaction.schema.ts
```

---

# 144. Form Validation Layers

Frontend Zod handles:

```text
required fields

basic amount validation

date format

enum selection
```

Backend handles everything again plus:

```text
ownership

category type

resource state

version

business conflict
```

---

# 145. Backend Validation Mapping

Utility:

```text
applyApiValidationErrors()
```

Concept:

```text
error.details.amount
↓
form.setError("amount", ...)
```

This should be reusable.

---

# 146. Dirty Form Protection

For complex forms like scenario builder:

```text
unsaved modifications
```

may warrant navigation confirmation.

For quick transaction form, keep flow lightweight.

---

# 147. Submit State

Buttons should:

```text
disable while submitting
```

and show:

```text
Saving...
```

or appropriate action.

This reduces accidental duplicate submissions.

---

# 148. Loading State Strategy

Three categories:

```text
Page Loading

Section Loading

Action Loading
```

Use the least disruptive state.

---

# 149. Page Loading

For first load:

```text
skeletons
```

matching page layout are preferred over a blank screen.

---

# 150. Section Loading

Dashboard may load a slow AI insight separately without blocking entire dashboard.

Example:

```text
Financial summary loaded
AI section loading
```

where endpoint architecture allows it.

If dashboard composite API returns everything together, use a coherent skeleton.

---

# 151. Action Loading

Examples:

```text
Creating transaction...

Calculating scenario...

Generating AI explanation...
```

Action labels should describe what is happening.

---

# 152. Empty States

Empty states must help user progress.

Bad:

```text
No data.
```

Preferred:

```text
No transactions yet.

Add your first income or expense to start understanding your cashflow.

[Add Transaction]
```

---

# 153. Account Empty State

```text
No accounts yet.

Add a bank account, e-wallet, savings account, or cash balance to get started.

[Add Account]
```

---

# 154. Budget Empty State

```text
No budgets yet.

Create a monthly category budget to monitor your spending.

[Create Budget]
```

---

# 155. Forecast Empty State

```text
Not enough financial context yet.

Add recurring income and expenses to improve your forecast.

[Add Recurring]
```

---

# 156. Scenario Empty State

```text
Test a financial decision before making it.

Examples:
- Buy a laptop
- Add an installment
- Reduce income
- Resign

[Create Scenario]
```

---

# 157. Insight Empty State

Avoid:

```text
Nothing detected.
```

Preferred:

```text
No new insights right now.

Savio will surface meaningful changes when it finds something worth your attention.
```

---

# 158. Error States

Error state should answer:

```text
what failed?

can user retry?

is existing data safe?
```

---

# 159. Forecast Error State

```text
Forecast could not be calculated.

Your financial records are safe.

[Try Again]
```

---

# 160. AI Error State

```text
Savio AI is temporarily unavailable.

Your financial calculations and records are unaffected.

[Try Again]
```

---

# 161. Inline vs Global Errors

Use inline errors for:

```text
forms

specific widgets
```

Use page-level states for:

```text
resource load failure
```

Use toast for:

```text
short-lived action feedback
```

---

# 162. Toast Strategy

Good toast examples:

```text
Transaction created.

Budget updated.

Scenario recalculated.

Session revoked.
```

Avoid using toast as the only place for important validation failures.

---

# 163. Confirmation Dialogs

Required for:

```text
void transaction

void transfer

end recurring rule

archive account

revoke session

archive scenario
```

---

# 164. Confirmation Content

A confirmation should explain consequence.

Bad:

```text
Are you sure?
```

Preferred:

```text
Void this transfer?

Rp1.000.000 will be returned to BCA Main and removed from GoPay.
```

---

# 165. Design System Philosophy

Savio should feel:

```text
calm

clear

trustworthy

modern

financially serious

not corporate-heavy

not crypto-like
```

---

# 166. Visual Direction

Avoid:

```text
neon trading UI

overly dark fintech dashboard

excessive gradients

gamification of financial risk
```

Prefer:

```text
clean surfaces

clear typography

generous spacing

subtle status emphasis

structured data hierarchy
```

---

# 167. Color Philosophy

Use neutral surfaces for most UI.

Semantic colors should communicate:

```text
positive

warning

negative

information
```

but not rely on color alone.

Exact palette belongs in `DESIGN.md`.

---

# 168. Typography

Use a modern readable sans-serif.

Typography hierarchy:

```text
Page Title

Section Title

Card Metric

Body

Supporting Text

Caption
```

Financial numbers should be highly legible.

---

# 169. Money Typography

Large financial numbers should use:

```text
tabular numerals
```

where supported.

This helps aligned values.

---

# 170. Shared UI Components

Recommended initial primitives:

```text
Button

Input

Textarea

Select

Combobox

Checkbox

Switch

Radio Group

Segmented Control

Dialog

Sheet

Drawer

Popover

Dropdown Menu

Tooltip

Tabs

Badge

Card

Table

Pagination

Skeleton

Alert

Toast

Progress

Empty State
```

---

# 171. Domain Components

Reusable domain components:

```text
MoneyValue

TransactionTypeBadge

AccountBadge

CategoryBadge

StatusBadge

BudgetProgress

GoalProgress

ForecastConfidenceBadge

ForecastEventBadge

InsightSeverityBadge

ScenarioComparisonMetric

FinancialMetricCard

DateRangePicker
```

---

# 172. MoneyValue Component

Concept:

```tsx
<MoneyValue
  amount="12000000.00"
  currency="IDR"
  sign="positive"
/>
```

This centralizes:

```text
formatting

sign

accessibility
```

---

# 173. Metric Card

Reusable:

```text
label

value

optional change

optional context

optional action
```

Examples:

```text
Income

Expense

Net Cashflow

Savings Rate
```

---

# 174. Status Badges

Central mapping:

```text
ACTIVE

PAUSED

ARCHIVED

POSTED

VOIDED

ON_TRACK

WARNING

EXCEEDED

LOW

MEDIUM

HIGH
```

Avoid feature-specific inconsistent styling for the same semantic status.

---

# 175. Data Table

Shared table should support:

```text
columns

loading

empty

actions

responsive behavior
```

But feature filters and business columns remain feature-owned.

---

# 176. Desktop Table / Mobile Card

Large tables such as transactions may render:

```text
table on desktop
```

and:

```text
compact list/card on mobile
```

rather than forcing tiny horizontal columns.

---

# 177. Accessibility

Minimum:

```text
keyboard navigation

focus states

semantic HTML

form labels

aria attributes where necessary

dialog focus trapping

color contrast

non-color status cues
```

---

# 178. Keyboard Navigation

Important flows should work without mouse.

Examples:

```text
transaction form

filters

dialogs

navigation
```

---

# 179. Focus Management

When opening modal:

```text
focus first meaningful field
```

When closing:

```text
return focus to trigger
```

---

# 180. Form Labels

Do not rely exclusively on placeholder.

Correct:

```text
Amount
[ Rp0 ]
```

Placeholder may supplement label.

---

# 181. Accessibility for Charts

Charts need accompanying textual summaries.

Example:

```text
Projected balance reaches its lowest point of Rp3.2M on 20 September.
```

---

# 182. Responsive Dashboard

Desktop:

```text
multi-column cards
```

Tablet:

```text
2 columns
```

Mobile:

```text
1 column
```

Priority order must remain meaningful.

---

# 183. Responsive Scenario Comparison

Desktop:

```text
Baseline | Scenario
```

Mobile:

```text
Metric
Baseline value
Scenario value
Difference
```

stacked.

---

# 184. Responsive Filters

Desktop:

```text
inline filter bar
```

Mobile:

```text
filter drawer
```

Active filters should still be visible through chips.

---

# 185. Active Filter Chips

Example:

```text
Expense ×

Food & Dining ×

Aug 2026 ×
```

Action:

```text
Clear all
```

---

# 186. Frontend Type Strategy

Prefer API-specific DTO types near feature API modules.

Do not expose raw backend response shape across the entire app if a small mapping layer improves clarity.

Example:

```text
TransactionResponse
↓
Transaction
```

may be identical initially.

Avoid unnecessary mapping if it adds no value.

---

# 187. Enum Types

Central domain enums may be shared where cross-feature.

Examples:

```text
TransactionType

AccountType

RecurringStatus

InsightSeverity
```

---

# 188. API Enum Safety

Frontend should still handle unknown server enum defensively.

Example:

```text
fallback badge:
Unknown
```

but TypeScript should model known values strictly.

---

# 189. Environment Configuration

Vite environment:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

Only public frontend configuration belongs in `VITE_*`.

Never expose:

```text
AI_API_KEY

DATABASE_URL

JWT_SECRET
```

to frontend.

---

# 190. Feature Flags

Frontend may receive backend capability configuration.

Example:

```text
AI enabled

notifications enabled

receipt upload enabled
```

Do not hardcode assumptions about optional infrastructure.

---

# 191. AI Capability State

Possible bootstrap response:

```json
{
  "ai": {
    "enabled": true,
    "available": false
  }
}
```

UI:

```text
AI feature exists
but provider currently unavailable
```

This is different from:

```text
AI disabled
```

---

# 192. Frontend Security Principles

Frontend security includes:

```text
no auth tokens in browser storage

no secret keys

safe rendering

no trust in route IDs

CSRF header

credentialed requests

minimal sensitive browser persistence
```

---

# 193. XSS Consideration

AI output and transaction descriptions are untrusted strings.

Render as text.

Do not use:

```text
dangerouslySetInnerHTML
```

for model output unless a hardened sanitizer is explicitly implemented.

---

# 194. AI Markdown

For MVP, AI responses may be rendered as:

```text
plain text
+
structured cards
```

instead of arbitrary model-generated Markdown/HTML.

This reduces rendering complexity and security risk.

---

# 195. Sensitive Data in Browser Logs

Avoid:

```text
console.log(apiResponse)
```

in production for financial endpoints.

Development logging should still avoid passwords/tokens.

---

# 196. Browser Cache Consideration

Authenticated financial API uses:

```text
Cache-Control: private/no-store
```

where appropriate.

TanStack Query still holds server state in memory for UX.

---

# 197. Query Stale Times

Not all data has equal freshness needs.

Examples:

```text
system categories
→ longer stale time

dashboard
→ shorter

transactions
→ moderate

forecast
→ explicit freshness from backend

AI insights
→ moderate
```

---

# 198. Forecast Freshness Is Domain State

Do not infer forecast freshness solely from TanStack Query cache.

Backend returns:

```text
FRESH

STALE
```

This domain state must be displayed.

---

# 199. Scenario Freshness Is Domain State

Likewise:

```text
scenario.is_stale
```

comes from backend.

Client cache freshness is separate.

---

# 200. URL State vs Local State

Use URL for:

```text
filters

sort

pagination

date ranges

selected analytics period
```

Use local state for:

```text
open dialog

temporary UI toggle

draft interaction
```

Use form state for:

```text
resource editing
```

---

# 201. Scenario Builder State

Scenario builder may combine:

```text
server-persisted scenario
+
local unsaved modification drafts
```

Keep distinction explicit.

---

# 202. Unsaved Scenario Modification

Possible UX:

```text
Add Change
↓
Draft Card
↓
Save Modification
```

or immediately persist after each modification.

For simpler implementation:

```text
persist modifications explicitly
```

then calculate server-side.

---

# 203. Scenario Calculation Trigger

Do not recalculate on every keystroke.

Use explicit:

```text
Calculate Scenario
```

Benefits:

```text
clear user intent

lower server load

predictable snapshots

better explainability
```

---

# 204. Forecast Calculation Trigger

Forecast page may calculate:

```text
on explicit horizon change
```

or load cached/latest snapshot.

Avoid creating persisted forecast snapshots on every rerender.

---

# 205. Dashboard Forecast Preview

Dashboard should generally consume:

```text
latest forecast preview
```

or a lightweight deterministic summary.

Do not trigger expensive long-horizon calculations unnecessarily.

---

# 206. AI Insight Generation UX

Automatic insights are background-generated.

Frontend should:

```text
read existing insights
```

not trigger new AI generation on every dashboard visit.

---

# 207. Manual AI Retry

If failed:

```text
Try Again
```

may issue a new AI request.

Do not auto-retry indefinitely.

---

# 208. AI Copilot Message Length

Composer should enforce client UX limit matching backend.

Example:

```text
4000 characters maximum
```

Exact value must align with API validation.

---

# 209. Copilot Pending State

After submit:

```text
Analyzing your financial data...
```

User may continue reading previous content.

Disable duplicate send until current request completes unless concurrency is intentionally supported.

---

# 210. Copilot Conversation MVP

For MVP, stateless/simple conversation is acceptable.

UI may display current session history in React state while page remains open.

Persistent server conversations can be P1.

---

# 211. Copilot Result Actions

Example:

```text
[View Food Transactions]

[Open Forecast]

[Create Scenario]
```

Actions come from backend allowlist.

---

# 212. Copilot Action Mapping

Frontend maintains mapping:

```text
VIEW_TRANSACTIONS
→ navigate("/transactions?...")

VIEW_FORECAST
→ navigate("/forecast")

OPEN_SCENARIO
→ navigate(...)
```

Do not execute arbitrary URLs from AI output.

---

# 213. Notification Deep Links

Notification backend provides:

```text
entity_type

entity_id
```

Frontend maps known entity types to routes.

Avoid trusting arbitrary server-generated external URL strings.

---

# 214. Breadcrumbs

Useful for detail pages:

```text
Accounts / BCA Main

Budgets / Food & Dining

Scenarios / Buy MacBook
```

Simple primary pages need not show excessive breadcrumbs.

---

# 215. Page Titles

Use concise titles:

```text
Dashboard

Transactions

Cashflow Forecast

Scenario Simulator

Savio Copilot
```

Avoid overly technical terminology in user-facing UI.

---

# 216. User-Friendly Domain Labels

Backend:

```text
INCOME_REDUCTION
```

UI:

```text
Reduce Income
```

Backend:

```text
RECURRING_EXPENSE
```

UI:

```text
Add Recurring Expense
```

Keep translation in frontend mapping.

---

# 217. Financial Language

Use:

```text
Projected

Estimated

Based on

Potential

Scenario
```

when discussing future state.

Avoid falsely certain language:

```text
Your balance will definitely be...
```

---

# 218. AI Language

Label AI output explicitly when useful:

```text
Savio AI Insight

Savio AI Explanation
```

But avoid placing `AI` on every sentence.

---

# 219. Deterministic vs AI UI Hierarchy

Recommended visual order:

```text
Financial Result

Supporting Facts

AI Explanation

User Actions
```

Never:

```text
AI Opinion
↓
hidden financial facts
```

---

# 220. Forecast Result Hierarchy

```text
Ending Balance

Lowest Balance

Timeline

Assumptions

AI Explanation if any
```

---

# 221. Scenario Result Hierarchy

```text
Baseline vs Scenario

Difference

Goal / Runway Impact

Assumptions

AI Explanation
```

---

# 222. Budget Result Hierarchy

```text
Budget

Actual

Remaining

Projected

Status

AI Explanation
```

---

# 223. Goal Result Hierarchy

```text
Target

Current

Required Contribution

Estimated Free Cashflow

Feasibility

Explanation
```

---

# 224. Frontend Testing Strategy

Testing layers:

```text
Unit

Component

Integration

E2E
```

---

# 225. Unit Tests

Suitable for:

```text
formatters

filter serialization

query key factories

validation schemas

status mapping

AI action mapping
```

---

# 226. Component Tests

Examples:

```text
TransactionForm

BudgetProgress

ScenarioComparison

AIInsightCard

ForecastConfidenceBadge
```

---

# 227. API Integration Tests

Using MSW:

```text
successful load

401 refresh

422 validation

409 conflict

429

500

AI unavailable
```

---

# 228. Authentication Tests

Critical:

```text
protected route loading

authenticated route

unauthenticated redirect

single refresh

refresh failure logout

no refresh loop
```

---

# 229. 401 Concurrency Test

Simulate three requests returning `401`.

Expected:

```text
one refresh request

three original retries after success
```

---

# 230. 409 Version Conflict Test

Example budget form:

```text
submit
→ 409 VERSION_CONFLICT
```

UI should:

```text
show conflict state

offer reload
```

not generic success/failure only.

---

# 231. Financial Form Tests

Transaction form:

```text
amount required

amount > 0

category required

backend validation mapping

submit disabled while pending
```

---

# 232. AI Categorization Tests

```text
suggestion appears

accept suggestion

reject/change suggestion

AI unavailable

invalid AI response handled by API
```

---

# 233. Forecast Tests

Test:

```text
summary metrics

event classification

stale state

low confidence

empty state

error state
```

---

# 234. Scenario Tests

Test:

```text
add modification

calculate

baseline comparison

difference

stale state

AI explanation unavailable

real finance state not represented as modified
```

---

# 235. Copilot Tests

Test:

```text
suggested prompt

answer

fact cards

clarification

rate limit

AI unavailable

allowed action mapping
```

---

# 236. E2E Critical Path

Playwright flow:

```text
Register
↓
Onboarding
↓
Create Account
↓
Add Income
↓
Add Expense
↓
Create Budget
↓
Create Goal
↓
Open Forecast
↓
Create Scenario
↓
Calculate Scenario
↓
Use Mock AI Copilot
↓
Logout
```

---

# 237. E2E Auth Path

Also test:

```text
login

expired access

successful refresh

logout

protected route redirect
```

---

# 238. Accessibility Testing

Use automated checks where possible plus manual keyboard review.

Examples:

```text
form labels

dialog focus

button names

contrast

keyboard navigation
```

---

# 239. Performance Considerations

Frontend performance priorities:

```text
route-level lazy loading

query deduplication

avoid unnecessary rerenders

avoid huge chart datasets

pagination

code splitting
```

---

# 240. Route-Level Lazy Loading

Large feature pages may use:

```text
React.lazy
```

or router-native lazy loading.

Examples:

```text
analytics

forecast

scenarios

copilot
```

---

# 241. Component Memoization

Do not use:

```text
useMemo

useCallback

React.memo
```

everywhere by default.

Optimize measurable rerender problems.

---

# 242. Chart Data Size

Forecast daily points for 12 months:

```text
~365 points
```

is acceptable.

Avoid rendering thousands of raw transactions as chart points unnecessarily.

---

# 243. Bundle Discipline

Do not add heavy libraries for functionality already covered by:

```text
browser APIs

small utilities

existing stack
```

---

# 244. Frontend Logging

Development errors may use console.

Production should prefer:

```text
central error reporting
```

if implemented.

Do not log sensitive financial payloads broadly.

---

# 245. Error Boundary

Use React error boundary around high-level application shell/routes.

Unexpected render failure should show:

```text
Something went wrong.

[Reload]
```

rather than blank page.

---

# 246. Query Error Boundary

Network/data errors should normally use feature-specific error state rather than crashing entire React tree.

---

# 247. 404 Page

Route not found:

```text
Page not found.

[Back to Dashboard]
```

---

# 248. Resource 404

If account/transaction does not exist:

```text
Transaction not found.
```

Do not reveal whether it exists for another user.

---

# 249. Frontend Naming Conventions

Files:

```text
kebab-case.tsx
```

Components:

```text
PascalCase
```

Functions/hooks:

```text
camelCase

useSomething
```

---

# 250. Component Responsibility

Prefer focused components.

Bad:

```text
DashboardPage.tsx
3000 lines
```

Preferred:

```text
DashboardPage

CashflowSummary

ForecastPreview

BudgetSummary

InsightPreview

GoalSummary
```

---

# 251. Avoid Excessive Fragmentation

Do not create one component per `<div>`.

Extract when:

```text
reusable

complex

independently testable

conceptually meaningful
```

---

# 252. Shared vs Feature Components

Example:

```text
Button
→ shared

MoneyValue
→ shared domain UI

TransactionTable
→ transactions feature

ScenarioComparison
→ scenarios feature
```

---

# 253. Frontend Data Mapping

API monetary strings remain strings through data layer.

Avoid:

```ts
Number("9999999999999999.99")
```

for authoritative calculation.

UI formatting may parse through a safe decimal library if needed.

---

# 254. Decimal Library

If frontend must perform display-only arithmetic:

```text
decimal.js
```

or equivalent may be used.

However, authoritative calculations stay backend-side.

---

# 255. Derived Display Values

Frontend may safely calculate purely visual things like:

```text
progress width
```

from backend-provided:

```text
utilization_percent
```

It should not recalculate utilization from raw transaction totals.

---

# 256. Currency Context

User default currency should be available globally for display.

However each account/amount response may include currency where ambiguity is possible.

---

# 257. Multi-Currency Future Proofing

Do not hardcode:

```text
Rp
```

inside components.

Always use:

```text
currency
```

parameter.

Even if MVP supports only IDR, this keeps formatting clean.

---

# 258. Locale Context

User locale may be provided through user profile.

Format:

```text
money

date

numbers
```

accordingly.

---

# 259. Theme

Initial implementation may use:

```text
light theme only
```

unless dark mode is intentionally included.

Do not spend excessive implementation time on theme infrastructure if not core to assessment.

---

# 260. Design Consistency

Every page should reuse:

```text
spacing

radius

typography

button variants

form styles

table behavior

status mapping
```

The product should look like one coherent application.

---

# 261. Loading Skeleton Consistency

Create shared skeleton primitives rather than random loaders across pages.

Examples:

```text
CardSkeleton

TableSkeleton

PageSkeleton
```

---

# 262. Dialog Consistency

Use shared confirmation dialog component with:

```text
title

description

confirm label

variant

pending state
```

---

# 263. Destructive Actions

Use clear destructive styling for:

```text
void

end recurring rule

archive where consequential

revoke session
```

But do not overuse red for ordinary navigation.

---

# 264. Onboarding Architecture

Steps:

```text
1. Welcome

2. Preferences

3. First Account

4. Optional Recurring Income

5. Complete
```

---

# 265. Onboarding State

Prefer backend-supported onboarding status or derive from authoritative account existence.

Avoid storing only:

```text
onboardingComplete=true
```

in localStorage.

---

# 266. Onboarding Progress

Show:

```text
Step 2 of 4
```

Simple progress, no unnecessary gamification.

---

# 267. Onboarding Skip

Optional steps may be skipped.

Required:

```text
at least enough data to enter app
```

If first account is required, backend/business flow should enforce that clearly.

---

# 268. Dashboard Onboarding Assistance

For new user, dashboard may show setup checklist:

```text
✓ Create Account

○ Add Income

○ Add Recurring Expenses

○ Create Budget

○ Create Goal
```

This should disappear naturally once setup is sufficient.

---

# 269. Product Education

Use short contextual explanations.

Example Forecast:

```text
Forecast combines recurring events and recent spending patterns to estimate future cashflow.
```

Avoid long tutorials blocking usage.

---

# 270. Explainable AI UX

AI explainability should be visible through:

```text
supporting facts

drivers

assumptions

source feature
```

not just through a generic AI icon.

---

# 271. AI Labels

Possible labels:

```text
Savio AI

AI Suggestion

AI Explanation

AI Insight
```

Keep consistent.

---

# 272. AI Confidence Labels

If numeric confidence is not calibrated, user-facing UI may convert:

```text
0.96
→ High Confidence
```

but methodology should be documented.

Alternative:

```text
96%
```

only if we can meaningfully explain what it represents.

---

# 273. No Fake Intelligence Animation

Avoid misleading:

```text
"AI is thinking through thousands of possibilities..."
```

if architecture is simply calling deterministic tools + one model.

Use precise UX:

```text
Analyzing your financial data...
```

---

# 274. Forecast vs Prediction Language

Prefer:

```text
Forecast

Projection

Estimate
```

not:

```text
Prediction guaranteed
```

---

# 275. Scenario vs Advice Language

Prefer:

```text
Impact

Trade-off

Comparison
```

rather than:

```text
Correct decision

Wrong decision
```

---

# 276. Cash Runway UX

If shown:

```text
Estimated runway:
4.1 months
```

with:

```text
Based on liquid balance and average essential monthly expenses.
```

---

# 277. Financial Health UX

If implemented P1, use:

```text
Cashflow Health
```

rather than broad:

```text
Financial Health
```

unless methodology supports it.

Display components, not score alone.

---

# 278. Search UX

Transaction search:

```text
Search transactions...
```

Debounce client request modestly.

Example:

```text
300 ms
```

Avoid request per keystroke without debounce.

---

# 279. Filter Reset

Always provide:

```text
Clear filters
```

when filters result in no records.

---

# 280. Pagination UX

Desktop:

```text
Previous  1 2 3 ... 8  Next
```

Mobile:

```text
Previous

Page 2 of 8

Next
```

---

# 281. Page Size

Default:

```text
20
```

Optional:

```text
20
50
100
```

Do not expose unlimited page size.

---

# 282. Client-Side Permission UI

Frontend may hide actions user cannot perform.

But backend remains authoritative.

Example:

```text
Viewer
→ no Edit button
```

Future household roles can use same architecture.

---

# 283. Current MVP Authorization UI

If roles are:

```text
USER

ADMIN
```

normal financial UI remains user-oriented.

Admin UI should be isolated from private finance unless specifically required.

---

# 284. Admin UI — Optional

If included for authorization demonstration:

```text
/admin
```

may contain:

```text
System Health

User Account Status

Worker Health

AI Provider Status
```

Not:

```text
all user transactions
```

by default.

---

# 285. Admin Route Guard

```text
role = ADMIN
```

required.

Frontend guard improves UX.

Backend permission remains final authority.

---

# 286. RBAC Future Compatibility

Navigation config may support:

```ts
requiredRoles
requiredPermissions
```

but avoid introducing a huge permission framework before it is needed.

---

# 287. Navigation Configuration

Example concept:

```ts
{
  label: 'Forecast',
  path: '/forecast',
  icon: TrendingUp
}
```

Central configuration allows:

```text
sidebar

mobile drawer

breadcrumbs
```

to reuse route metadata.

---

# 288. Route Metadata

Could include:

```text
title

navigation group

permission

breadcrumb label
```

Do not overcomplicate route registry.

---

# 289. Frontend Acceptance Criteria

The frontend architecture is implemented correctly when:

```text
routes are protected

auth bootstrap does not flicker incorrectly

no auth tokens stored in local/session storage

Axios refresh is single-flight

401 does not cause infinite retry

403 does not force logout

422 maps to forms

429 is handled

500 is safe

all major pages have loading states

all list pages have empty states

financial mutations have pending states

critical actions have confirmation

financial calculations come from backend

AI is presented as interpretation

responsive navigation works

mobile layouts remain usable

components are reusable without becoming generic abstractions
```

---

# 290. Dashboard Acceptance Criteria

Dashboard must provide:

```text
total balance

income

expense

net cashflow

forecast preview

budget status

goal status

recent insights

upcoming financial events
```

with:

```text
loading

empty

error

responsive
```

states.

---

# 291. Transaction Acceptance Criteria

Must support:

```text
list

search

filter

sort

pagination

create income

create expense

AI category suggestion

edit

void

validation

version conflict
```

---

# 292. Forecast Acceptance Criteria

Must support:

```text
horizon selection

summary metrics

projected chart

event timeline

confidence

assumptions

stale state

recalculate

responsive layout
```

---

# 293. Scenario Acceptance Criteria

Must support:

```text
create scenario

multiple modifications

calculate

baseline vs scenario

difference metrics

goal impact

chart

assumptions

stale state

recalculate

AI explanation

mobile layout
```

---

# 294. Copilot Acceptance Criteria

Must support:

```text
message input

suggested prompts

structured facts

clarification

scenario handoff

allowed actions

AI error

rate limit

no arbitrary frontend action execution
```

---

# 295. Critical Frontend Demo Flow

Recommended final demo:

```text
Login
    ↓
Dashboard
    ↓
Quick Add Expense
    ↓
AI Category Suggestion
    ↓
Transaction Created
    ↓
Budget Updates
    ↓
AI Insight Appears
    ↓
Forecast
    ↓
Scenario Simulator
    ↓
Buy Laptop Scenario
    ↓
Baseline vs Scenario
    ↓
AI Explanation
    ↓
Copilot Question
```

This shows a cohesive product rather than disconnected pages.

---

# 296. Critical Frontend Failure Demo

Useful review scenarios:

```text
401
→ refresh
→ retry

refresh failure
→ login

422
→ inline form errors

409
→ conflict warning

AI unavailable
→ finance still usable

forecast stale
→ recalculate

scenario stale
→ recalculate

empty account
→ actionable empty state
```

---

# 297. Frontend Architecture Trade-Off — TanStack Query

Decision:

```text
TanStack Query for server state
```

Reason:

```text
Savio is server-data heavy.
```

Benefits:

```text
cache

loading state

error state

deduplication

invalidation

mutation handling
```

---

# 298. Frontend Architecture Trade-Off — No Large Global Store Initially

Decision:

```text
no Redux/Zustand required initially
```

Reason:

```text
most important state is server state or local form/UI state
```

A global store can be introduced later if real cross-feature client state appears.

---

# 299. Frontend Architecture Trade-Off — React Hook Form

Decision:

```text
React Hook Form
```

because:

```text
many forms

good performance

clear server error mapping

strong Zod integration
```

---

# 300. Frontend Architecture Trade-Off — URL Filters

Decision:

```text
list/filter state in URL
```

because:

```text
refresh-safe

back-button friendly

debuggable

shareable
```

---

# 301. Frontend Architecture Trade-Off — Minimal Optimistic Financial Updates

Decision:

```text
prefer authoritative refetch after critical finance mutation
```

Reason:

```text
financial correctness > instant illusion
```

---

# 302. Frontend Architecture Trade-Off — AI Structured UI

Decision:

```text
render structured AI responses
```

not only chat bubbles.

Reason:

```text
trust

explainability

actionability

validation
```

---

# 303. Frontend Architecture Trade-Off — Explicit Scenario Calculate

Decision:

```text
calculate scenario on user action
```

rather than on every form change.

Reason:

```text
clear snapshots

lower load

better mental model
```

---

# 304. Frontend Architecture Trade-Off — Light Theme First

Decision:

```text
prioritize one polished theme
```

before implementing theme switching.

Reason:

```text
UI quality matters more than feature count.
```

Dark mode may be added only if time permits.

---

# 305. Frontend Architecture Trade-Off — Modal vs Page

Use modal/sheet for:

```text
quick create

simple edit

confirmation
```

Use dedicated page for:

```text
forecast

scenario

analytics

complex detail
```

This keeps frequent actions fast while preserving space for complex workflows.

---

# 306. Frontend Architecture Source-of-Truth Hierarchy

The UI must preserve:

```text
BACKEND AUTHORITY
       ↓
DETERMINISTIC FINANCIAL RESULT
       ↓
FRONTEND PRESENTATION
       ↓
OPTIONAL AI EXPLANATION
       ↓
USER ACTION
```

The frontend must not invert this into:

```text
UI assumption
↓
financial truth
```

---

# 307. Final Frontend Architecture Model

```text
                         SAVIO FRONTEND
                               │
                               ▼
                       Application Router
                               │
                               ▼
                        Application Shell
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
      Money                 Planning             Intelligence
        │                      │                      │
 Accounts                Budgets                 Insights
 Transactions            Goals                   Copilot
 Recurring               Forecast
                         Scenarios
        │                      │                      │
        └──────────────────────┼──────────────────────┘
                               ▼
                         Feature Modules
                               │
                               ▼
                        TanStack Query
                               │
                               ▼
                          Axios Client
                               │
                               ▼
                         Savio REST API
                               │
                               ▼
                      Deterministic Backend
```

---

# 308. Final User Experience Model

The interface should guide users through:

```text
RECORD
   ↓
UNDERSTAND
   ↓
PLAN
   ↓
FORECAST
   ↓
SIMULATE
   ↓
EXPLAIN
   ↓
DECIDE
```

Each major screen should reinforce that progression.

---

# 309. Final Frontend Principle

Savio should not look or behave like:

```text
a CRUD admin dashboard
```

and it should not become:

```text
a chatbot with financial tables attached
```

The intended experience is:

```text
structured personal finance application
+
deterministic financial intelligence
+
clear planning tools
+
AI-assisted explanation
```

The user should always be able to distinguish:

```text
what actually happened

what Savio calculated

what Savio estimates

what a scenario assumes

what AI is explaining
```

The final Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**