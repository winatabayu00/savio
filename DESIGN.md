# Savio — Design System & UI/UX Guidelines

## Related Documents

- [README.md](README.md) — project overview, setup, and documentation index.
- [Product Foundation](docs/product/product-foundation.md) — product vision, positioning, and target users.
- [Frontend Architecture](docs/architecture/frontend-architecture.md) — routes, feature organization, and UI architecture.
- [User Flows](docs/product/user-flows.md) — end-to-end user workflows.

## 1. Document Purpose

This document defines the visual design system, interaction principles, layout rules, component behavior, information hierarchy, responsive behavior, and user experience standards for Savio.

The purpose of this document is to ensure that the Savio interface feels:

- coherent,
- trustworthy,
- modern,
- financially serious,
- easy to understand,
- and clearly differentiated from a generic CRUD dashboard.

This document should be used as the visual and UX source of truth during frontend implementation.

The product principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

The UI must visually preserve that hierarchy.

---

# 2. Product Design Positioning

Savio is not:

```text
a trading terminal
a crypto dashboard
a banking backoffice
a generic admin panel
a chatbot-first product
```

Savio is:

> **A calm, explainable personal cashflow intelligence and financial decision support experience.**

The interface should make financial information easier to understand without creating unnecessary visual pressure.

---

# 3. Core Design Principles

Savio follows these design principles:

```text
CLARITY
TRUST
CONTEXT
CALMNESS
EXPLAINABILITY
CONTROL
CONSISTENCY
```

---

# 4. Clarity

Financial information must be easy to scan.

Users should quickly understand:

```text
What is my current position?

What changed?

What is projected?

What is hypothetical?

What is AI-generated?
```

Important numbers should never be hidden inside large text blocks.

---

# 5. Trust

Trust is critical because Savio handles personal financial information.

The interface should avoid:

```text
flashy animations
gamified financial pressure
neon trading colors
fake urgency
aggressive upselling
```

Prefer:

```text
clear labels
visible assumptions
stable layouts
predictable interactions
precise numbers
explainable states
```

---

# 6. Context

A number without context has limited value.

Avoid:

```text
Expenses
Rp8.400.000
```

Prefer:

```text
Expenses
Rp8.400.000

+23.5% vs recent baseline
```

or:

```text
Expenses
Rp8.400.000

Rp1.6M above your recent average
```

where supported by deterministic backend data.

---

# 7. Calmness

Financial products can create anxiety.

Savio should communicate risks without making the interface feel alarming.

Use:

```text
Potential cashflow pressure
```

instead of:

```text
DANGER! YOU WILL RUN OUT OF MONEY!
```

Use appropriate severity, but avoid sensational language.

---

# 8. Explainability

Important decisions should expose:

```text
Result
Facts
Drivers
Assumptions
AI Explanation
```

The UI should make clear why Savio is showing a particular conclusion.

---

# 9. User Control

The UI must reinforce that:

```text
AI suggests
Savio calculates
User decides
```

Users should be able to:

```text
review
edit
dismiss
recalculate
accept
decline
```

where appropriate.

---

# 10. Consistency

The same concepts must look and behave consistently.

Examples:

```text
ACTIVE
```

should not be green text on one page and blue badge on another.

```text
Rp1.000.000
```

should follow one formatting rule everywhere.

---

# 11. Visual Personality

The intended Savio personality is:

```text
professional
calm
intelligent
clear
warm
modern
minimal
```

Avoid looking:

```text
corporate-heavy
playful-fintech
crypto-inspired
trading-oriented
overly futuristic
AI-gimmicky
```

---

# 12. Brand Concept

Product name:

```text
Savio
```

Meaning in product context:

```text
a financial intelligence companion
focused on understanding,
planning,
and decision support
```

Primary product statement:

> **Understand your money today. See what comes next. Test decisions before making them.**

---

# 13. Logo Direction

Initial logo can remain simple.

Preferred concept:

```text
Savio wordmark
+
minimal abstract mark
```

Potential visual ideas:

```text
flow
path
signal
financial trajectory
forward movement
layered insight
```

Avoid:

```text
coins
dollar signs
candlestick charts
bank buildings
robot heads
generic AI sparkles as the primary logo
```

---

# 14. Design System Foundation

The design system consists of:

```text
Color
Typography
Spacing
Radius
Shadow
Border
Iconography
Motion
Components
Layouts
States
```

---

# 15. Color Philosophy

Use a restrained palette.

Primary purposes:

```text
Brand
Neutral
Positive
Warning
Negative
Information
AI
```

The product should rely mostly on neutral surfaces.

Semantic colors should be used only where meaningful.

---

# 16. Recommended Color Direction

Suggested palette direction:

```text
Primary
Deep Indigo / Blue

Neutral
Slate / Zinc

Positive
Green

Warning
Amber

Negative
Red

Information
Blue

AI Accent
Violet / Indigo
```

This is a direction rather than a strict final hex palette.

---

# 17. Recommended Initial Tokens

Suggested Tailwind-compatible semantic tokens:

```text
background
surface
surface-muted
border
text-primary
text-secondary
text-muted

primary
primary-hover
primary-subtle

success
success-subtle

warning
warning-subtle

danger
danger-subtle

info
info-subtle

ai
ai-subtle
```

---

# 18. Initial Light Theme Palette

Suggested starting values:

```text
Background:
#F8FAFC

Surface:
#FFFFFF

Surface Muted:
#F1F5F9

Border:
#E2E8F0

Text Primary:
#0F172A

Text Secondary:
#475569

Text Muted:
#64748B

Primary:
#4F46E5

Primary Hover:
#4338CA

Primary Subtle:
#EEF2FF

Success:
#15803D

Success Subtle:
#F0FDF4

Warning:
#B45309

Warning Subtle:
#FFFBEB

Danger:
#B91C1C

Danger Subtle:
#FEF2F2

Info:
#0369A1

Info Subtle:
#F0F9FF

AI:
#7C3AED

AI Subtle:
#F5F3FF
```

These values may be adjusted during implementation if visual testing reveals contrast or hierarchy problems.

---

# 19. Color Usage Rules

Primary color:

```text
main actions
active navigation
selected controls
focus accents
```

Success:

```text
positive cashflow
goal achieved
healthy state
```

Warning:

```text
budget approaching limit
forecast concern
medium-risk status
```

Danger:

```text
budget exceeded
negative balance
destructive action
high-risk state
```

AI:

```text
AI insight
AI suggestion
AI explanation
Copilot identity
```

Do not use AI violet for ordinary financial facts.

---

# 20. Income and Expense Colors

Recommended:

```text
Income
→ positive semantic color

Expense
→ neutral/danger-adjacent treatment
```

Do not make every expense bright red.

Ordinary expense records are normal financial activity, not necessarily errors.

A possible style:

```text
Income:
green text

Expense:
default text with minus sign

Negative financial state:
red
```

---

# 21. Color Accessibility

Never communicate state using color alone.

Example:

Bad:

```text
green dot
```

Preferred:

```text
green dot
+
ON TRACK
```

---

# 22. Typography

Recommended font direction:

```text
Inter
```

or another highly legible modern sans-serif.

The final product should use one primary UI typeface.

---

# 23. Typography Scale

Suggested:

```text
Display
32px / 40px

Page Title
28px / 36px

Section Title
20px / 28px

Card Title
16px / 24px

Body
14–16px / 22–24px

Small
13px / 20px

Caption
12px / 18px
```

---

# 24. Typography Weights

Use primarily:

```text
400
500
600
700
```

Avoid excessive ultra-bold typography.

---

# 25. Financial Number Typography

Important financial values should use:

```text
font-weight: 600–700
font-variant-numeric: tabular-nums
```

Example:

```text
Rp16.250.000
```

This improves readability and alignment.

---

# 26. Money Hierarchy

Primary financial metric:

```text
Rp16.250.000
```

Supporting label:

```text
Total Balance
```

Supporting comparison:

```text
+Rp3.6M net this month
```

The number should remain visually dominant.

---

# 27. Spacing System

Use a predictable spacing scale.

Suggested:

```text
4
8
12
16
20
24
32
40
48
64
```

Tailwind's default spacing can support this.

---

# 28. Page Spacing

Desktop content padding:

```text
24–32px
```

Mobile:

```text
16px
```

Large dashboard sections:

```text
24–32px vertical spacing
```

---

# 29. Component Internal Spacing

Cards:

```text
16–24px
```

Forms:

```text
16–20px
```

Table cells:

```text
12–16px
```

---

# 30. Border Radius

Recommended:

```text
Small:
6px

Default:
8px

Card:
12px

Large:
16px
```

Avoid excessive pill shapes across the whole product.

---

# 31. Buttons

Button radius:

```text
8px
```

Pill-style buttons should only be used where semantically appropriate, such as filter chips.

---

# 32. Shadows

Use subtle shadows.

Recommended:

```text
cards mostly border-based

elevated overlays:
small soft shadow
```

Avoid:

```text
large floating shadows on every card
```

---

# 33. Borders

Borders are important for calm visual separation.

Use:

```text
1px neutral border
```

for:

```text
cards
inputs
tables
dialogs
```

---

# 34. Iconography

Recommended icon set:

```text
Lucide
```

Reasons:

```text
consistent
simple
modern
broad coverage
```

---

# 35. Icon Style

Use:

```text
outline icons
```

with consistent size.

Typical sizes:

```text
16px
18px
20px
24px
```

---

# 36. Motion

Motion should be subtle.

Recommended:

```text
150–250ms
```

for:

```text
hover
drawer
dialog
dropdown
accordion
```

Avoid financial-value animations that make numbers feel unstable.

---

# 37. Reduced Motion

Respect:

```text
prefers-reduced-motion
```

for significant UI transitions.

---

# 38. Application Shell

Desktop:

```text
┌───────────────┬─────────────────────────────────────────────┐
│               │ Header                                      │
│ Sidebar       ├─────────────────────────────────────────────┤
│               │                                             │
│               │ Main Content                                │
│               │                                             │
└───────────────┴─────────────────────────────────────────────┘
```

---

# 39. Sidebar Dimensions

Recommended expanded width:

```text
256px
```

Collapsed:

```text
72px
```

if collapse is implemented.

---

# 40. Sidebar Layout

```text
Savio

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

---

# 41. Sidebar Section Labels

Section labels should be visually subtle.

Example:

```text
OVERVIEW
```

Use:

```text
small
muted
medium weight
```

Navigation items should remain more prominent.

---

# 42. Active Navigation

Active item:

```text
primary subtle background
primary text
active icon
```

Avoid:

```text
full saturated primary rectangle
```

unless it still feels calm.

---

# 43. Sidebar Logo Area

Top:

```text
Savio
```

with optional small mark.

Height:

```text
64–72px
```

aligning with header.

---

# 44. Header

Recommended height:

```text
64px
```

Contains:

```text
page title or breadcrumb
quick add
notifications
copilot shortcut
user menu
```

---

# 45. Quick Add Button

Primary header action:

```text
+ Add
```

Click:

```text
Income
Expense
Transfer
Recurring
Budget
Goal
```

Most-used options should appear first.

---

# 46. User Menu

Contains:

```text
Profile

Settings

Sessions

Logout
```

Avoid too many unrelated actions.

---

# 47. Main Content Width

Data-heavy pages:

```text
full available width
```

Settings/forms:

```text
max-width around 800–1000px
```

Scenario / dashboard:

```text
wide responsive layout
```

---

# 48. Dashboard Layout

Desktop:

```text
┌──────────────────────────────────────────────────────────────┐
│ Greeting / Financial Context                                │
├─────────────┬─────────────┬─────────────┬───────────────────┤
│ Balance     │ Income      │ Expense     │ Net Cashflow      │
├────────────────────────────────────┬─────────────────────────┤
│ Cashflow Trend                     │ Forecast Preview        │
├────────────────────────────────────┼─────────────────────────┤
│ Budgets                            │ Goals                   │
├────────────────────────────────────┴─────────────────────────┤
│ Savio Insights                                               │
├──────────────────────────────────────────────────────────────┤
│ Upcoming Activity                                            │
└──────────────────────────────────────────────────────────────┘
```

---

# 49. Dashboard Greeting

Avoid generic:

```text
Welcome back!
```

without value.

Prefer:

```text
Good evening, Alex

Here is how your cashflow looks this month.
```

The greeting should not dominate the screen.

---

# 50. Financial Summary Cards

Each card:

```text
Label

Primary Value

Supporting Comparison

Optional Icon
```

Example:

```text
Net Cashflow

Rp3.600.000

30% of income retained
```

---

# 51. Card Rules

Cards should avoid:

```text
too many borders
too many nested boxes
```

One card should usually answer one primary question.

---

# 52. Metric Card Variants

Variants:

```text
default

positive

warning

danger

AI
```

Semantic variant should be subtle.

Example warning card:

```text
neutral card
+
small amber status
```

not full orange background.

---

# 53. Chart Visual Style

Charts should use:

```text
clean axes
minimal grid lines
limited colors
clear tooltip
```

Avoid:

```text
3D charts
donut overload
rainbow category charts
```

---

# 54. Cashflow Chart

Recommended:

```text
line or area chart
```

for balance/cashflow trend.

If showing:

```text
income vs expense
```

use:

```text
bar chart
```

where comparison is clearer.

---

# 55. Category Distribution

Prefer:

```text
horizontal bars
```

for many categories.

Donut may be acceptable for a small number of top-level categories, but should not be the default for large lists.

---

# 56. Chart Tooltip

Tooltip example:

```text
24 Aug 2026

Balance
Rp8.250.000

Event
Rent -Rp3.000.000
```

---

# 57. Tables

Desktop tables should feel lightweight.

Recommended row height:

```text
52–60px
```

Use:

```text
subtle row border
hover state
```

Avoid heavy grid borders around every cell.

---

# 58. Transaction Table

Columns:

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

---

# 59. Table Header

Use:

```text
small-medium typography
muted text
subtle background if needed
```

Sticky headers may be added for long lists.

---

# 60. Table Row Interaction

Clicking row may open:

```text
transaction detail sheet
```

Do not make every cell separately clickable.

---

# 61. Mobile Transactions

Mobile list card:

```text
GrabFood
Food & Dining

24 Aug · GoPay

-Rp87.500
```

Optional secondary:

```text
POSTED
```

---

# 62. Forms

Forms should use:

```text
clear labels

supporting descriptions where necessary

inline validation

consistent vertical rhythm
```

---

# 63. Form Layout

Simple forms:

```text
single column
```

Complex forms on desktop may use:

```text
two-column grouping
```

only when fields naturally belong together.

---

# 64. Input Height

Recommended:

```text
40–44px
```

for standard inputs.

Touch targets on mobile should remain at least:

```text
44px
```

---

# 65. Input State

States:

```text
default

hover

focus

disabled

error

success where useful
```

---

# 66. Focus Style

Use visible focus ring:

```text
primary color
```

Do not remove browser focus behavior without replacement.

---

# 67. Error Field

Example:

```text
Amount

[ -100 ]

Amount must be greater than zero.
```

Error message:

```text
small
danger
directly below field
```

---

# 68. Form Helper Text

Use for important context:

```text
Current balance is calculated from your financial activity and cannot be edited directly.
```

---

# 69. Select / Combobox

Use combobox for longer lists such as:

```text
categories
accounts
```

especially when custom categories grow.

---

# 70. Financial Amount Input

Example display:

```text
Rp
[ 1.500.000 ]
```

Internally, maintain safe decimal representation.

---

# 71. Transaction Form

Layout:

```text
Type

Amount

Account

Category

Date

Merchant / Description

Notes
```

Primary fields should be above the fold.

---

# 72. Transaction Type Selector

Use segmented buttons:

```text
Income
Expense
```

Transfer is separate.

---

# 73. AI Category Suggestion

Suggested layout:

```text
┌─────────────────────────────────────────┐
│ ✦ Savio AI Suggestion                   │
│                                         │
│ Food & Dining                           │
│ High confidence                         │
│                                         │
│ [Use Category]                          │
└─────────────────────────────────────────┘
```

Use subtle AI accent.

---

# 74. AI Suggestion Distinction

Do not style suggestion exactly like confirmed field state.

The user must understand:

```text
suggestion ≠ saved data
```

---

# 75. Accounts Page

Preferred card layout.

Example:

```text
┌───────────────────────────────────┐
│ BCA Main                          │
│ Bank                              │
│                                   │
│ Rp14.750.000                      │
│                                   │
│ Active                            │
└───────────────────────────────────┘
```

---

# 76. Account Type Icons

Possible:

```text
Cash
wallet icon

Bank
landmark / building icon

E-Wallet
smartphone icon

Savings
piggy-bank icon
```

Icons are supporting, not authoritative.

---

# 77. Reconciliation Design

Reconciliation deserves intentional UX because it changes tracked balance.

Layout:

```text
Tracked Balance
Rp4.800.000

Actual Balance
[ Rp5.000.000 ]

Difference
+Rp200.000
```

Then:

```text
Reason
```

and confirmation.

---

# 78. Recurring Page

Recommended view:

```text
Upcoming

Active Recurring

Paused
```

or simple filter tabs.

---

# 79. Recurring Timeline

Useful future pattern:

```text
01 Sep
Rent

10 Sep
Internet

12 Sep
Netflix

25 Sep
Salary
```

This can make future cashflow more intuitive.

---

# 80. Budget Design

Budget card:

```text
Food & Dining

Rp1.450.000
of Rp2.000.000

██████████████░░░░░ 72.5%

Rp550.000 remaining

ON TRACK
```

---

# 81. Budget Warning

At warning threshold:

```text
WARNING

82% used
```

Use amber semantic treatment.

---

# 82. Budget Exceeded

Example:

```text
EXCEEDED

Rp250.000 over budget
```

Use danger styling but remain calm.

---

# 83. Projected Budget State

Must clearly label:

```text
Projected
```

Example:

```text
Projected month-end spend

Rp2.350.000
```

Do not visually mix this with actual spend.

---

# 84. Goal Card

Example:

```text
Emergency Fund

Rp12.000.000 / Rp30.000.000

40%

Target
Feb 2027

Required
Rp3M / month

AT RISK
```

---

# 85. Goal Progress

Use progress bar with target.

If goal is achieved:

```text
ACHIEVED
```

with positive semantic state.

Avoid celebratory confetti unless intentionally justified.

---

# 86. Analytics Design

Analytics should prioritize:

```text
summary
then explanation
then charts
```

Suggested hierarchy:

```text
Period

Key Metrics

Period Comparison

Expense Breakdown

Largest Changes
```

---

# 87. Period Comparison Card

Example:

```text
Expenses increased

+Rp1.600.000
+23.5%

vs 3-month average
```

Drivers:

```text
Food +Rp700k

Shopping +Rp550k
```

---

# 88. Forecast Visual Identity

Forecast is deterministic but uncertain.

Visually distinguish forecast data from historical data.

Recommended:

```text
historical line:
solid

forecast line:
dashed
```

or distinct but related treatment.

---

# 89. Forecast Page Layout

```text
Header

Horizon Selector

Summary Cards

Balance Projection

Timeline

Assumptions

Confidence
```

---

# 90. Forecast Summary Cards

Recommended:

```text
Projected End Balance

Lowest Balance

Projected Income

Projected Expense
```

Confidence as a separate contextual card/badge.

---

# 91. Forecast Timeline Event

Example:

```text
10 Sep

Internet

-Rp450.000

Scheduled
```

Badge:

```text
Recurring
```

---

# 92. Forecast Event Styles

User-facing labels:

```text
Confirmed
Recurring
Estimated
Assumption
```

Backend enum mapping:

```text
KNOWN
SCHEDULED
ESTIMATED
ASSUMED
```

---

# 93. Forecast Assumptions Panel

Example:

```text
Forecast assumptions

• Variable spending estimated using recent average.
• Monthly salary expected on the 25th.
• Current recurring expenses remain unchanged.
```

---

# 94. Forecast Confidence

Example:

```text
Medium confidence
```

Supporting explanation:

```text
Based on 3 months of financial history.
```

Use an info icon for methodology details.

---

# 95. Low Balance Alert

Example:

```text
Potential cashflow pressure

Your projected balance reaches Rp850.000 on 18 Sep.

Next expected salary:
25 Sep

[Explore Scenario]
```

---

# 96. Scenario Simulator Visual Identity

Scenario Simulator should feel like:

```text
a decision workspace
```

not:

```text
a form followed by a JSON result
```

It is one of Savio's most visually important screens.

---

# 97. Scenario Desktop Layout

```text
┌───────────────────────────────┬──────────────────────────────┐
│ Scenario Builder              │ Comparison                   │
│                               │                              │
│ Buy MacBook                   │ Baseline     Scenario        │
│ 6 months                      │                              │
│                               │ Ending balance               │
│ Changes                       │ Rp18.4M       Rp3.4M          │
│                               │                              │
│ - MacBook Rp15M               │ Lowest balance               │
│                               │ Rp8.2M        Rp1.1M          │
│ [+ Add Change]                │                              │
│                               │ Cash runway                  │
│ [Calculate Scenario]          │ 4.1 mo        1.9 mo          │
└───────────────────────────────┴──────────────────────────────┘
```

Below:

```text
Projection Chart

Goal Impact

Assumptions

Savio AI Explanation
```

---

# 98. Scenario Builder Panel

Panel should contain:

```text
Scenario Name

Horizon

Changes
```

Each modification shown as a compact card.

---

# 99. Scenario Change Card

Example:

```text
One-Time Expense

MacBook Pro

Rp15.000.000

10 Sep 2026

[Edit] [Remove]
```

---

# 100. Add Scenario Change

Button:

```text
+ Add Change
```

Opens selector:

```text
Spend Once

Receive Income Once

Add Recurring Expense

Add Recurring Income

Reduce Income

Stop Income

Reduce Expense

Change Savings
```

Use human-friendly language.

---

# 101. Scenario Comparison Table

Recommended:

```text
Metric             Baseline       Scenario      Difference

Ending Balance     Rp18.4M        Rp3.4M        -Rp15M

Lowest Balance     Rp8.2M         Rp1.1M        -Rp7.1M

Savings Rate       27%            8%            -19 pts

Cash Runway        4.1 mo         1.9 mo         -2.2 mo
```

---

# 102. Scenario Difference Emphasis

Difference column should be visually prominent enough to scan.

Do not use only red/green.

Include:

```text
- Rp15M
```

or:

```text
+ 3 months delay
```

---

# 103. Scenario Chart

Show:

```text
Baseline
Scenario
```

on one timeline.

Use two clearly distinguishable line styles.

Avoid more than 2–3 scenario lines on one chart in MVP.

---

# 104. Scenario Goal Impact

Example:

```text
Emergency Fund

Baseline:
Dec 2026

Scenario:
Mar 2027

Delay:
3 months
```

---

# 105. Scenario Stale State

Prominent but non-destructive banner:

```text
This scenario is out of date.

Your financial data changed after the last calculation.

[Recalculate]
```

---

# 106. AI Scenario Explanation Design

Below deterministic result:

```text
┌─────────────────────────────────────────────────┐
│ ✦ Savio AI                                     │
│                                                 │
│ This purchase does not immediately create a    │
│ negative balance, but it materially reduces    │
│ your cash buffer.                              │
│                                                 │
│ Key impact                                     │
│ • Lowest balance falls to Rp1.1M               │
│ • Runway decreases by 2.2 months               │
│ • Emergency fund delayed by 3 months           │
└─────────────────────────────────────────────────┘
```

---

# 107. AI Card Style

AI cards should use:

```text
subtle violet/indigo accent
neutral background
small AI icon
```

Do not use:

```text
glowing gradient border
animated sparkles everywhere
```

---

# 108. AI Insights Feed

Insight cards should have:

```text
severity
title
short explanation
main fact
action
```

---

# 109. Insight Severity Design

Example:

```text
INFO
blue/neutral

LOW
neutral/info

MEDIUM
amber

HIGH
red
```

AI purple identifies the source as AI, not severity.

---

# 110. Insight Card Example

```text
MEDIUM

Dining spending increased

Dining is 60% above your recent baseline.

Main driver
Food delivery +Rp720k

[View Details]
```

AI label may appear subtly:

```text
Savio AI
```

---

# 111. Insight Detail

Suggested structure:

```text
Header
Severity + Type

What Changed

Supporting Facts

Drivers

Savio AI Explanation

Suggested Actions

Feedback
```

---

# 112. Supporting Facts

Example:

```text
Current
Rp2.4M

Baseline
Rp1.5M

Change
+60%
```

These are deterministic.

---

# 113. AI Copilot Visual Direction

Copilot should feel integrated with Savio, not a separate ChatGPT clone.

Use:

```text
financial context cards
structured actions
natural-language explanation
```

---

# 114. Copilot Layout

Desktop:

```text
┌──────────────────────────────────────────────────────────┐
│ Savio Copilot                                            │
├──────────────────────────────────────────────────────────┤
│ Suggested prompts                                        │
│                                                          │
│ Conversation                                             │
│                                                          │
│ Financial fact cards                                     │
│                                                          │
├──────────────────────────────────────────────────────────┤
│ Ask about your cashflow...                         Send   │
└──────────────────────────────────────────────────────────┘
```

---

# 115. Copilot Suggested Prompts

Use chips/cards:

```text
Why did I spend more this month?

Which budget is at risk?

What are my largest recurring expenses?

Can I afford a Rp10M purchase?

What happens if I resign?
```

---

# 116. Copilot Message Styling

User:

```text
simple primary-subtle bubble
```

Savio:

```text
neutral content block
```

Avoid excessive chat bubble nesting when structured facts exist.

---

# 117. Copilot Fact Card

Example:

```text
Current Expense
Rp8.4M

Recent Baseline
Rp6.8M

Difference
+Rp1.6M
```

---

# 118. Copilot Actions

Example:

```text
[View Food Transactions]

[Open Forecast]

[Create Scenario]
```

Use normal Savio buttons.

Do not create visually mysterious AI-only actions.

---

# 119. Copilot Clarification

Example:

```text
When should I assume the purchase happens?

[Today]

[Next Payday]

[Choose Date]
```

Structured interaction is preferred to ambiguous free-form responses.

---

# 120. AI Loading State

Use:

```text
Analyzing your financial data...
```

for Copilot.

For AI categorization:

```text
Finding a category suggestion...
```

For scenario explanation:

```text
Preparing an explanation...
```

---

# 121. AI Failure State

Example:

```text
Savio AI is temporarily unavailable.

Your financial calculations and records are unaffected.

[Try Again]
```

---

# 122. AI Disabled State

If user disables AI:

```text
AI insights are disabled.

You can enable them again in Settings.
```

Do not show provider error styling.

---

# 123. Notification Design

Header:

```text
bell icon
```

Unread:

```text
small count badge
```

---

# 124. Notification Popover

Shows up to:

```text
5 recent notifications
```

Footer:

```text
View all notifications
```

---

# 125. Notification Item

Layout:

```text
Icon

Title

Short message

Time
```

Unread item may have subtle background.

---

# 126. Settings Design

Settings should use left sub-navigation on desktop.

```text
Profile

Preferences

Security

Sessions
```

Mobile:

```text
stacked cards / tabs
```

---

# 127. Settings Form Width

Recommended:

```text
640–760px
```

Do not stretch simple settings across full dashboard width.

---

# 128. Security Settings Visual Style

Security actions should be direct and serious.

Examples:

```text
Active Sessions

Revoke Session

Logout All Devices
```

Destructive actions need confirmation.

---

# 129. Loading States

Savio should use skeletons for initial content loading.

Examples:

```text
MetricCardSkeleton

ChartSkeleton

TableSkeleton

InsightSkeleton
```

---

# 130. Loading Principle

Avoid replacing the whole page with:

```text
one giant spinner
```

when the layout is known.

Skeleton should preserve expected structure.

---

# 131. Button Loading

Example:

```text
Save
```

becomes:

```text
Saving...
```

with spinner.

Button remains disabled.

---

# 132. Forecast Loading

Use:

```text
Calculating forecast...
```

if computation is triggered explicitly.

For persisted forecast read:

```text
Loading forecast...
```

---

# 133. Scenario Loading

Use:

```text
Calculating scenario...
```

Optional secondary text:

```text
Comparing your baseline and scenario.
```

---

# 134. Empty States

Every data collection needs intentional empty state.

---

# 135. Transactions Empty State

```text
No transactions yet.

Add your first income or expense to start understanding your cashflow.

[Add Transaction]
```

---

# 136. Budget Empty State

```text
No budgets yet.

Create a category budget to track how your spending compares with your plan.

[Create Budget]
```

---

# 137. Goals Empty State

```text
No financial goals yet.

Create a goal and Savio will help you track its progress.

[Create Goal]
```

---

# 138. Forecast Empty State

```text
Not enough financial context yet.

Adding recurring income and expenses will improve your forecast.

[Add Recurring]
```

---

# 139. Scenario Empty State

```text
Test a financial decision before making it.

Try scenarios like:
• Buying a laptop
• Adding an installment
• Losing income
• Increasing savings

[Create Scenario]
```

---

# 140. AI Insights Empty State

```text
No new insights right now.

Savio will surface meaningful changes when something deserves your attention.
```

---

# 141. Error State

Error state includes:

```text
title
description
retry where safe
```

---

# 142. Generic Page Error

```text
Unable to load this page.

Please try again.

[Retry]
```

---

# 143. Financial Error

Example:

```text
Unable to update this transaction.

No financial changes were applied.

[Try Again]
```

Only claim no changes if backend contract guarantees rollback.

---

# 144. Conflict State

For:

```text
409 VERSION_CONFLICT
```

use dedicated UX:

```text
This record has changed since you opened it.

Reload the latest version before continuing.

[Reload]
```

---

# 145. Rate Limit State

Example:

```text
Too many requests.

Please wait 30 seconds before trying again.
```

Especially for:

```text
Copilot
login
```

---

# 146. Toasts

Success:

```text
Transaction created.

Budget updated.

Scenario recalculated.
```

Failure:

Use toast for non-form/general actions only.

Validation should remain inline.

---

# 147. Toast Duration

Typical:

```text
3–5 seconds
```

Longer only for meaningful warnings.

---

# 148. Dialog Design

Dialog structure:

```text
Title

Description

Content

Cancel

Primary / Destructive Action
```

---

# 149. Destructive Confirmation

Example:

```text
Void this transaction?

The original financial effect will be reversed and your account balance will be recalculated.

Reason
[                      ]

[Cancel] [Void Transaction]
```

---

# 150. Archive Confirmation

Archive is less destructive than delete.

Use:

```text
neutral warning
```

not extreme danger styling.

---

# 151. Drawers and Sheets

Use side sheet for:

```text
quick transaction detail

filters

lightweight edit
```

Use dialog for:

```text
confirmation

small forms
```

Use full page for:

```text
forecast

scenario

analytics
```

---

# 152. Mobile Bottom Sheet

Good for:

```text
filters

quick action selection

transaction detail
```

Ensure swipe/close behavior remains accessible through explicit button.

---

# 153. Status Badge Design

Base badge:

```text
small text
rounded 6–8px
subtle background
```

Avoid huge pills.

---

# 154. Financial Status Mapping

Examples:

```text
ACTIVE
neutral/positive

PAUSED
neutral

ARCHIVED
muted

POSTED
neutral/positive

VOIDED
muted/danger

ON_TRACK
success

WARNING
warning

EXCEEDED
danger

LOW
neutral/info

MEDIUM
warning

HIGH
danger
```

---

# 155. Accessibility Contrast

All text and interactive elements should meet WCAG AA where practical.

Minimum normal text contrast target:

```text
4.5:1
```

Large text:

```text
3:1
```

---

# 156. Focus Accessibility

Interactive elements must have visible keyboard focus.

Example:

```text
2px primary ring
```

---

# 157. Keyboard Interaction

Expected:

```text
Tab navigation

Enter / Space activation

Escape closes modal

Arrow keys where component conventions support them
```

---

# 158. Dialog Accessibility

Dialogs require:

```text
focus trap

aria-labelledby

aria-describedby where appropriate

focus return on close
```

---

# 159. Icon Button Accessibility

Icon-only button must have:

```text
aria-label
```

Example:

```text
aria-label="Open notifications"
```

---

# 160. Chart Accessibility

Provide:

```text
text summary

meaningful tooltip

legend

keyboard access where library supports it
```

---

# 161. Responsive Strategy

Savio is designed mobile-aware from the beginning.

Layouts should transform intentionally.

---

# 162. Mobile Application Shell

Recommended:

```text
Top Header

Page Content

Bottom Navigation
```

---

# 163. Mobile Bottom Navigation

Potential:

```text
Home

Transactions

Forecast

Copilot

More
```

---

# 164. Mobile Quick Add

Potential central/prominent:

```text
+
```

or header action.

Avoid overlapping important bottom nav items.

---

# 165. Mobile Dashboard

Order:

```text
Balance

Income / Expense

Net Cashflow

Forecast

Budget

Goals

Insights

Upcoming
```

One column.

---

# 166. Mobile Metrics

Cards may use:

```text
2-column mini grid
```

where values remain readable.

Example:

```text
Income | Expense
```

---

# 167. Mobile Scenario

Use:

```text
builder
↓
calculate
↓
comparison
↓
chart
↓
goal impact
↓
AI explanation
```

---

# 168. Mobile Comparison

Instead of wide table:

```text
Ending Balance

Baseline
Rp18.4M

Scenario
Rp3.4M

Difference
-Rp15M
```

repeat by metric.

---

# 169. Mobile Tables

Do not render an 8-column table squeezed into 360px.

Transform into:

```text
compact cards/list
```

or controlled horizontal scroll if necessary.

---

# 170. Tablet Layout

Tablet may use:

```text
collapsed sidebar
```

or drawer.

Dashboard:

```text
2-column grid
```

where appropriate.

---

# 171. Breakpoint Strategy

Suggested:

```text
< 640
mobile

640–1023
tablet

>= 1024
desktop
```

Exact Tailwind breakpoints may be used.

---

# 172. Page-Level Responsive Rule

Each page must explicitly verify:

```text
mobile

tablet

desktop
```

during implementation.

Responsive behavior is not complete merely because CSS does not overflow.

---

# 173. Search and Filter Design

Transaction search input:

```text
Search transactions...
```

Leading search icon.

---

# 174. Desktop Filters

Example:

```text
[Search] [Type] [Account] [Category] [Date] [More Filters]
```

---

# 175. Mobile Filters

```text
[Search]

[Filters 3]
```

Opens drawer.

---

# 176. Active Filter Chips

Example:

```text
Expense ×

Food & Dining ×

This Month ×
```

---

# 177. Clear Filters

Always available when filters active:

```text
Clear all
```

---

# 178. Pagination Design

Desktop:

```text
Showing 1–20 of 142

Previous  1 2 3 ... 8  Next
```

---

# 179. Mobile Pagination

```text
Previous

Page 2 of 8

Next
```

---

# 180. Date Range Picker

Should support:

```text
preset periods

custom date range
```

Presets:

```text
This Month

Last Month

Last 3 Months

This Year
```

---

# 181. Financial Data Formatting

Use locale-aware presentation.

For Indonesian:

```text
Rp1.500.000
```

Dates:

```text
24 Agu 2026
```

or chosen English UI equivalent if interface is English.

The implementation should choose one consistent application language.

---

# 182. Language Strategy

Recommended for take-home:

```text
English interface
```

with:

```text
IDR formatting
```

Reason:

```text
technical review/readability
```

However, Indonesian interface is also valid if implemented consistently.

Do not mix languages unpredictably.

---

# 183. Copywriting Principle

Use plain language.

Prefer:

```text
Add Expense
```

over:

```text
Create Expense Transaction Record
```

Prefer:

```text
Cashflow Forecast
```

over overly technical prediction terminology.

---

# 184. Confirmation Copy

Be explicit about consequence.

Example:

```text
Pause this recurring expense?

Future occurrences will no longer be posted until you resume it.
```

---

# 185. Error Copy

Avoid blaming user.

Bad:

```text
You entered invalid data.
```

Preferred:

```text
Please check the highlighted fields.
```

---

# 186. AI Copy

Avoid overclaiming.

Bad:

```text
Savio knows exactly what will happen.
```

Preferred:

```text
Based on your current financial data and assumptions...
```

---

# 187. Forecast Copy

Use:

```text
Projected

Estimated

Expected

Assumed
```

correctly.

---

# 188. Scenario Copy

Use:

```text
If this scenario happens...
```

not:

```text
When this happens...
```

---

# 189. Empty State Illustration

Optional minimal illustration may be used.

Do not require custom illustration library.

Simple icon + text is sufficient and often cleaner.

---

# 190. AI Iconography

Use one consistent symbol for Savio AI.

Possible:

```text
sparkles
brain-like abstract icon
wand
```

Use sparingly.

---

# 191. AI vs System Notification

AI insight:

```text
AI icon
```

System warning:

```text
alert icon
```

Do not visually conflate the two.

---

# 192. Loading AI Icon

Subtle animated AI icon may be acceptable.

Keep motion restrained.

---

# 193. Onboarding Design

Onboarding should feel lightweight.

Desktop:

```text
centered card
```

with progress.

---

# 194. Onboarding Step Layout

```text
Step 1 of 4

Set up your financial preferences

Timezone
Currency

[Back] [Continue]
```

---

# 195. Onboarding Welcome

Example:

```text
Welcome to Savio

Understand where your money goes,
see what comes next,
and test decisions before making them.

[Get Started]
```

---

# 196. Onboarding First Account

Make account creation simple:

```text
Account name

Type

Starting balance
```

Optional institution.

---

# 197. Onboarding Recurring Income

Example:

```text
Do you receive regular income?
```

Options:

```text
Yes, add salary/income

Skip for now
```

---

# 198. Onboarding Completion

Example:

```text
You're ready to use Savio.

Your dashboard will become more useful as you add financial activity.

[Go to Dashboard]
```

---

# 199. Setup Checklist

Dashboard may show temporary checklist:

```text
Set up Savio

✓ Create your first account
○ Add income
○ Add recurring expenses
○ Create a budget
○ Create a financial goal
```

---

# 200. Checklist Dismissal

Allow dismiss once user understands.

Do not permanently consume dashboard space.

---

# 201. Design Tokens

Recommended token categories:

```text
colors

font

font size

font weight

line height

spacing

radius

shadow

z-index

transition
```

---

# 202. Tailwind Token Strategy

Prefer semantic tokens through CSS variables.

Example:

```css
:root {
  --background: ...;
  --foreground: ...;
  --primary: ...;
  --border: ...;
}
```

Tailwind can reference them.

This avoids scattering raw hex values.

---

# 203. No Random Colors

Avoid:

```text
bg-blue-500
bg-purple-400
text-red-600
```

chosen arbitrarily across features.

Use semantic component variants.

---

# 204. Z-Index Scale

Suggested:

```text
base
dropdown
sticky
overlay
modal
toast
```

Avoid random:

```text
z-[999999]
```

---

# 205. Layering

Recommended concept:

```text
base        0
sticky      20
dropdown    40
overlay     50
modal       60
toast       70
```

Exact values may vary.

---

# 206. Component Variants

Button:

```text
primary

secondary

outline

ghost

danger
```

Badge:

```text
neutral

success

warning

danger

info

ai
```

Alert:

```text
info

success

warning

danger
```

---

# 207. Button Sizes

```text
sm

md

lg
```

Default:

```text
md
```

Avoid excessive size variants.

---

# 208. Primary Action Rule

Each screen should ideally have one clearly dominant primary action.

Example transactions:

```text
Add Transaction
```

Secondary actions remain less visually prominent.

---

# 209. Destructive Button Rule

Danger variant only for destructive action confirmation.

Do not use danger button for ordinary navigation.

---

# 210. Card Variants

Card types:

```text
surface

metric

interactive

alert

AI
```

Avoid dozens of card styles.

---

# 211. Interactive Card

Hover:

```text
subtle border/background change
```

Cursor indicates clickability.

Static cards should not have click hover effects.

---

# 212. Progress Bars

Used for:

```text
budgets

goals
```

Not for:

```text
forecast certainty
```

unless methodology supports numeric progress.

---

# 213. Skeletons

Skeletons should roughly match:

```text
actual component dimensions
```

Do not create visually unrelated loading placeholders.

---

# 214. Tooltip

Use for:

```text
technical financial concept

forecast confidence methodology

AI source explanation
```

Do not hide essential information exclusively in tooltip.

---

# 215. Popover

Use for:

```text
notifications

date quick filters

small contextual details
```

---

# 216. Dropdown Menu

Use for:

```text
row actions

account actions

user menu
```

Do not put primary actions only inside hidden menus.

---

# 217. Tabs

Use when sections represent sibling views.

Examples:

```text
Account:
Overview | Transactions

Recurring:
Active | Paused

Insights:
All | Cashflow | Budget
```

Avoid excessive nested tabs.

---

# 218. Breadcrumb

Use for detail hierarchy.

Example:

```text
Scenarios / Buy MacBook
```

Not necessary on dashboard.

---

# 219. Page Header

Pattern:

```text
Title

Short description

Primary action
```

Example:

```text
Cashflow Forecast

See how your current financial patterns may affect future balances.

[Recalculate]
```

---

# 220. Description Length

Keep page descriptions:

```text
1–2 short lines
```

Do not turn headers into documentation sections.

---

# 221. AI Explainability Hierarchy

Important AI display should follow:

```text
1. Deterministic fact
2. AI explanation
3. Suggested action
```

Example:

```text
Dining +60% vs baseline

Savio AI:
Food delivery was the largest contributor.

[Review Transactions]
```

---

# 222. Human Decision Hierarchy

Scenario:

```text
Calculation

Interpretation

Options

User Decision
```

Do not display:

```text
Recommended: BUY / DON'T BUY
```

as final binary AI authority.

---

# 223. Alternative Scenarios

Potential design:

```text
Explore Alternatives

Buy Today

Buy Next Month

Save Rp5M First
```

These can be cards.

AI may propose alternatives, but deterministic engine recalculates them.

---

# 224. Scenario Presets

Potential useful shortcuts:

```text
Large Purchase

Income Reduction

Resignation

New Installment

Increase Savings
```

These start scenario templates.

---

# 225. Forecast Presets

```text
30 Days

60 Days

90 Days

6 Months

12 Months
```

Use segmented control or select on mobile.

---

# 226. AI Suggested Prompt Cards

Use small clickable cards:

```text
Why did I spend more?

Can I afford a large purchase?

Which budget needs attention?
```

Do not show dozens.

---

# 227. Data Density

Savio should balance:

```text
financial detail
+
visual breathing room
```

Avoid dashboard screens with:

```text
20 tiny metric cards
```

---

# 228. Dashboard Priority

Recommended:

```text
4 primary financial metrics maximum in first row
```

Other values appear contextually later.

---

# 229. Dense Data Pages

Transactions and reports may be denser than dashboard.

Use page-specific density intentionally.

---

# 230. Desktop Max Width

Dashboard may use:

```text
max-width: 1600px
```

or full responsive width.

Forms may use smaller max width.

---

# 231. Fixed vs Fluid

Sidebar:

```text
fixed
```

Header:

```text
sticky/fixed
```

Content:

```text
fluid
```

depending on implementation.

---

# 232. Sticky Table Header

Recommended when table scrolls vertically on desktop.

---

# 233. Sticky Scenario Builder

On wide desktop, scenario builder may remain sticky while comparison scrolls.

Only if it improves usability and does not cause layout complexity.

---

# 234. Mobile Sticky Actions

For long forms, primary action may be in sticky bottom bar.

Example:

```text
[Calculate Scenario]
```

Ensure it does not cover content.

---

# 235. Financial Risk Communication

Use severity carefully.

Example:

```text
Low projected balance
```

with:

```text
warning
```

Only use danger if:

```text
negative projected balance
```

or clearly severe deterministic threshold.

---

# 236. Positive Trends

Savio should also surface positive financial changes.

Example:

```text
Savings rate improved

30%
+8 percentage points
```

This prevents AI/analytics from feeling purely negative.

---

# 237. Positive Insight Style

Use subtle success treatment.

Avoid gamified celebration.

---

# 238. AI Insight Severity vs Sentiment

Severity:

```text
financial importance
```

Sentiment:

```text
positive / negative
```

These are different.

A positive trend may still use:

```text
INFO
```

severity.

---

# 239. Technical Error Visuals

Do not show raw:

```text
500
422
409
```

as primary user copy.

Translate to user-friendly message.

Technical code may be shown only in small debug/reference details where useful.

---

# 240. Request ID Display

Example:

```text
Reference: req_abc123
```

small muted text on unexpected error.

---

# 241. Privacy UX

Avoid casually displaying excessive financial details in notifications or public-facing contexts.

In-app notification is acceptable.

Future email/push notifications may require privacy-aware copy.

---

# 242. Session Security UX

Example:

```text
MacBook · Chrome
Current session

Jakarta, Indonesia
Last active now
```

Location should only be shown if safely derived and accurate enough.

Do not overclaim precise location.

---

# 243. Authentication Pages

Login/register should be visually simpler than application shell.

Layout:

```text
Savio logo

Form

Minimal product message
```

---

# 244. Login Page

Example:

```text
Welcome back

Continue to Savio

Email
Password

[Sign In]

Don't have an account?
Create one
```

---

# 245. Register Page

Example:

```text
Create your Savio account

Name

Email

Password

Confirm Password

[Create Account]
```

---

# 246. Auth Layout

Desktop may use:

```text
form side

subtle product statement side
```

but avoid large marketing-heavy hero if unnecessary.

---

# 247. Password Field

Support:

```text
show/hide
```

button with accessible label.

---

# 248. Password Requirements

Show concise requirements during registration.

Example:

```text
At least 8 characters
```

or match actual backend policy.

Frontend and backend requirements must remain aligned.

---

# 249. Logout UX

Logout does not need confirmation in most cases.

Logout all sessions should confirm.

---

# 250. Notification Preferences

If implemented:

```text
Budget Alerts

Upcoming Bills

Cashflow Risk

AI Insights
```

Use switches.

---

# 251. AI Preferences Copy

Example:

```text
AI Insights

Allow Savio to generate explanations from your financial patterns.

Core financial calculations remain available when disabled.
```

---

# 252. Data Quality Communication

Forecast and AI should communicate when data is limited.

Example:

```text
Limited history

Savio currently has 18 days of transaction data.
```

---

# 253. Data Quality Badge

Possible:

```text
Limited Data
```

with info tooltip.

Do not style as an error unless it blocks a calculation.

---

# 254. Source Indicators

For projected events:

```text
Recurring

Estimated

User Assumption
```

This improves traceability.

---

# 255. Estimated Value Style

Use subtle visual treatment:

```text
dashed underline
or
label
```

to distinguish from confirmed values.

Avoid reducing contrast so far that values become unreadable.

---

# 256. Historical vs Future

Chart:

```text
historical
solid

future projection
dashed
```

with:

```text
Today
```

vertical marker.

---

# 257. Today Marker

Forecast chart should visually show:

```text
Today
```

where historical becomes projected.

---

# 258. Scenario Marker

Scenario chart can mark:

```text
Purchase Date
```

or:

```text
Income stops
```

to explain divergence.

---

# 259. Hover Tooltips on Scenario Chart

Example:

```text
10 Sep

Baseline
Rp13.2M

Scenario
-Rp1.8M

MacBook Purchase
-Rp15M
```

---

# 260. Financial Goal Scenario Impact

Display:

```text
Goal delayed by 3 months
```

not only dates.

Both can appear:

```text
Dec 2026 → Mar 2027
```

---

# 261. Reports Design

Reports should resemble analytical pages rather than printable documents in MVP.

Sections:

```text
Summary

Trend

Category Breakdown

Budget Performance

Goal Progress
```

---

# 262. Report Export

If export is added:

```text
Export
```

secondary action.

Do not prioritize over core analytics.

---

# 263. MinIO / Receipt Future UX

Receipt attachment should appear as:

```text
Attachment
```

within transaction detail.

Upload status:

```text
Uploading
Processing
Ready
Failed
```

---

# 264. Receipt AI Extraction UX

Future:

```text
Savio extracted:

Merchant
Amount
Date
Category Suggestion

Review before creating transaction.
```

Never use:

```text
Transaction created automatically
```

without clear policy/user confirmation.

---

# 265. Import UX

Future staged interface:

```text
Upload
↓
Validate
↓
Review
↓
Confirm
```

Show:

```text
Valid rows

Warnings

Invalid rows
```

---

# 266. Design Review Checklist

For every page, verify:

```text
What is the primary user question?

What is the primary action?

What information is authoritative?

What information is projected?

What information is AI-generated?

Is loading handled?

Is empty handled?

Is error handled?

Is mobile usable?

Is keyboard navigation usable?

Is important state understandable without color?
```

---

# 267. Component Review Checklist

For every shared component:

```text
Does it have one clear responsibility?

Is it actually reused?

Does it support loading/disabled state?

Is it accessible?

Does it support responsive use?

Does it rely on semantic tokens?
```

---

# 268. Financial UX Review Checklist

For financial values:

```text
Is currency shown?

Is sign/direction clear?

Is actual vs projected clear?

Is historical vs scenario clear?

Are authoritative numbers from backend?

Is precision preserved?
```

---

# 269. AI UX Review Checklist

For AI features:

```text
Is AI clearly identified?

Are deterministic facts visible?

Can the user reject the suggestion?

Can AI failure be handled?

Is AI copy non-authoritative?

Are actions allowlisted?

Is sensitive data minimized?
```

---

# 270. Responsive Review Checklist

Verify at least:

```text
375px

768px

1024px

1440px
```

for important screens.

---

# 271. Critical Screens

Highest visual-quality priority:

```text
Login

Onboarding

Dashboard

Transactions

Forecast

Scenario Simulator

AI Insights

Savio Copilot
```

---

# 272. Secondary Screens

Still polished but lower complexity:

```text
Accounts

Recurring

Budgets

Goals

Settings

Notifications
```

---

# 273. Design Implementation Priority

## P0

```text
Design tokens

Application shell

Navigation

Buttons

Inputs

Cards

Tables

Dialogs

Forms

Loading

Empty

Error

Dashboard

Transactions

Budgets

Goals

Forecast

Scenario

AI Insights

Copilot

Responsive behavior
```

---

## P1

```text
Notification center

Advanced analytics

Session management polish

Additional charts

Setup checklist

Advanced AI structured actions
```

---

## P2

```text
Dark mode

Advanced animations

Receipt workflows

Import UI

Advanced report export

Household UX
```

---

# 274. Design Anti-Patterns

Do not:

```text
copy a generic admin template unchanged

put every feature in a card

use gradients everywhere

use 10 different colors on one dashboard

hide critical data behind hover

use AI as the first visual hierarchy

make scenario results only textual

display forecast as guaranteed truth

use meaningless charts

use placeholder lorem ipsum in final product

leave empty pages with "No data"

rely on desktop-only tables

show raw API error responses
```

---

# 275. Dashboard Anti-Pattern

Avoid:

```text
20 KPI cards
+
6 charts
+
3 tables
```

The dashboard should prioritize understanding.

---

# 276. AI Anti-Pattern

Avoid:

```text
Ask Savio anything!
```

as the primary product experience.

Copilot is a supporting interaction layer.

---

# 277. Scenario Anti-Pattern

Avoid:

```text
Form
↓
Raw JSON
```

The comparison should be visual and understandable.

---

# 278. Forecast Anti-Pattern

Avoid showing only:

```text
Projected balance: Rp12M
```

without:

```text
timeline
assumptions
confidence
event sources
```

---

# 279. Transaction Anti-Pattern

Avoid requiring:

```text
5 separate screens
```

to add a simple expense.

Frequent flows should remain fast.

---

# 280. Design Definition of Done

A UI feature is complete when:

```text
happy state

loading state

empty state

error state

pending state

validation state

responsive state

accessible interaction

API error handling

consistent design tokens
```

are implemented.

---

# 281. Critical Demo Design Flow

The final visual demo should feel like one connected experience.

```text
Login
↓
Dashboard

Current balance
Income
Expense
Forecast
Insight
↓
Quick Add Expense

GrabFood
Rp250k
↓
AI Suggestion

Food & Dining
↓
Budget updates
↓
Insight

Dining increased
↓
Forecast

Future cashflow
↓
Scenario

Buy MacBook Rp15M
↓
Comparison

Baseline vs Scenario
↓
AI Explanation
↓
Copilot

"What is the biggest impact?"
```

---

# 282. Demo Visual Narrative

The reviewer should visually see the Savio thesis:

```text
TRACK
↓
UNDERSTAND
↓
PREDICT
↓
SIMULATE
↓
EXPLAIN
↓
DECIDE
```

without requiring a long verbal explanation.

---

# 283. Design System Source-of-Truth Hierarchy

The design should reflect:

```text
ACTUAL FINANCIAL DATA
        ↓
DETERMINISTIC ANALYSIS
        ↓
FORECAST / SCENARIO
        ↓
AI EXPLANATION
        ↓
USER ACTION
```

Visual hierarchy must not invert these levels.

---

# 284. Primary Visual Hierarchy

In most intelligence screens:

```text
1. Financial Result
2. Context
3. Explanation
4. Action
```

---

# 285. Forecast Visual Hierarchy

```text
Projected Result
↓
Timeline
↓
Assumptions
↓
Confidence
```

---

# 286. Scenario Visual Hierarchy

```text
Baseline vs Scenario
↓
Difference
↓
Impact
↓
Assumptions
↓
AI Explanation
↓
Explore Alternative
```

---

# 287. AI Insight Visual Hierarchy

```text
Signal
↓
Facts
↓
AI Explanation
↓
Suggested Action
```

---

# 288. Final Design Philosophy

Savio should make financial information feel:

```text
less fragmented

less abstract

less reactive
```

and more:

```text
understandable

forward-looking

explainable

intentional
```

The interface should help users move from:

> **"What happened to my money?"**

to:

> **"Why did it happen?"**

then:

> **"What may happen next?"**

then:

> **"What changes if I make this decision?"**

and finally:

> **"I understand the trade-offs well enough to decide."**

---

# 289. Final Design Rule

The UI should always make it possible to distinguish:

```text
ACTUAL
PROJECTED
HYPOTHETICAL
AI-GENERATED
```

financial information.

If the user cannot tell those apart, the design is incorrect.

---

# 290. Final Product Design Principle

Savio is not designed to impress users with financial complexity.

It is designed to reduce that complexity.

The visual system must therefore remain:

```text
calm
clear
structured
responsive
trustworthy
explainable
```

And the final hierarchy remains:

> **Finance Engine calculates. AI interprets. User decides.**