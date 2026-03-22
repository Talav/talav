# Talav Framework Documentation

> **Living documentation for the Talav Go Framework - optimized for human-LLM collaboration**

---

## What is Talav?

Talav is a modular Go framework for building production-ready HTTP APIs with:
- **Enforced layered architecture** (components → FX modules → application)
- **Type-safe HTTP handlers** via Zorya API framework
- **Automatic OpenAPI generation** from struct tags
- **RFC-compliant error handling** (RFC 9457)
- **Declarative route security** (auth, roles, permissions)

**Core Value**: Ship type-safe REST APIs **3-5x faster** with clean architecture that scales.

---

## Documentation Structure

```
/docs
├── 00-context/              # WHY and WHAT EXISTS NOW
│   ├── vision.md            # Product vision, goals, principles
│   ├── system-map.md        # Current architecture and components
│   └── assumptions.md       # Risks, unknowns, assumptions
│
├── 01-product/              # WHAT the product must do
│   └── prd.md               # Product requirements document
│
├── 02-features/             # HOW features are designed & built
│   └── (future feature specs go here)
│
├── 03-logs/                 # MEMORY (decisions, changes, learnings)
│   ├── decision-log.md      # Why we made architectural choices
│   ├── implementation-log.md # What changed in code and why
│   ├── bug-log.md           # Bugs, fixes, root causes
│   └── insights.md          # Patterns and learnings
│
├── 04-process/              # HOW to work with this system
│   ├── dev-workflow.md      # Daily development loop
│   └── llm-prompts.md       # AI assistant prompt templates
│
└── README.md                # This file (navigation hub)
```

---

## Quick Start Guides

### For New Developers

1. **Understand WHY** → Read [vision.md](00-context/vision.md) (10 min)
2. **See WHAT exists** → Read [system-map.md](00-context/system-map.md) (20 min)
3. **Learn HOW to work** → Read [dev-workflow.md](04-process/dev-workflow.md) (15 min)
4. **Start building** → Use [llm-prompts.md](04-process/llm-prompts.md) with AI assistant

**Total onboarding**: ~45 minutes to productive contribution

---

### For AI Assistants (LLMs)

**Context to Always Load**:
1. [.cursor/rules/](../.cursor/rules/) - Architecture constraints (MUST READ)
2. [vision.md](00-context/vision.md) - Product principles
3. [system-map.md](00-context/system-map.md) - Component structure
4. [dev-workflow.md](04-process/dev-workflow.md) - Development patterns

**Key Constraints** (from `.cursor/rules/`):
- Components (`pkg/component/*`) CANNOT import `pkg/fx/*` or `pkg/module/*`
- FX modules (`pkg/fx/*`) CANNOT import `pkg/module/*`
- Application (`pkg/module/*`, `main.go`) can import everything

**Prompt Templates**: See [llm-prompts.md](04-process/llm-prompts.md) for canonical prompts

---

## Documentation Philosophy

This system follows **"Docs That Remember"** principles:

1. **Logs Over Perfect Docs**
   - Chronological logs never go stale
   - Capture reality (what happened) not aspiration (what should be)
   - Decision log, implementation log, bug log, insights

2. **Just-in-Time Documentation**
   - **Before**: Vision, requirements, design
   - **During**: Decisions, implementation notes
   - **After**: Outcomes, bugs, learnings

3. **Single Source of Truth**
   - Each type of information has ONE place
   - Requirements → PRD
   - Decisions → Decision log
   - Reality → Implementation log
   - Learnings → Insights

4. **AI-Native Structure**
   - Clear templates LLMs can follow
   - Explicit context in each doc
   - Architecture rules in machine-readable format
   - Canonical prompts that work

---

## How to Use This Documentation

### Planning a Feature

1. Check [prd.md](01-product/prd.md) - Is it already prioritized?
2. Review [assumptions.md](00-context/assumptions.md) - Any related unknowns?
3. Consult [decision-log.md](03-logs/decision-log.md) - Past decisions that apply?
4. Create feature spec in `02-features/` (when ready)

### Building a Feature

1. Follow [dev-workflow.md](04-process/dev-workflow.md) - Daily development loop
2. Use [llm-prompts.md](04-process/llm-prompts.md) - AI assistant prompts
3. Reference [system-map.md](00-context/system-map.md) - Component structure
4. Respect [.cursor/rules/](../.cursor/rules/) - Architecture constraints

### Debugging an Issue

1. Check [bug-log.md](03-logs/bug-log.md) - Similar bug fixed before?
2. Review [implementation-log.md](03-logs/implementation-log.md) - Recent changes?
3. Consult [insights.md](03-logs/insights.md) - Known patterns?

### Making a Decision

1. Document in [decision-log.md](03-logs/decision-log.md) - Context, options, rationale
2. After implementation, update with actual outcome
3. Extract learnings to [insights.md](03-logs/insights.md)

---

## Key Documents by Role

### For Framework Users

- **Getting Started**: [vision.md](00-context/vision.md) → [dev-workflow.md](04-process/dev-workflow.md)
- **API Reference**: Component READMEs in `pkg/component/*/README.md`
- **Examples**: See `examples/` directory (TBD)
- **Troubleshooting**: [bug-log.md](03-logs/bug-log.md)

### For Contributors

- **Architecture**: [system-map.md](00-context/system-map.md) + [.cursor/rules/](../.cursor/rules/)
- **Workflow**: [dev-workflow.md](04-process/dev-workflow.md)
- **Patterns**: [insights.md](03-logs/insights.md)
- **History**: [implementation-log.md](03-logs/implementation-log.md)

### For LLM Assistants

- **Context**: [vision.md](00-context/vision.md) + [system-map.md](00-context/system-map.md)
- **Constraints**: [.cursor/rules/](../.cursor/rules/) (CRITICAL - always read first)
- **Prompts**: [llm-prompts.md](04-process/llm-prompts.md)
- **Patterns**: [insights.md](03-logs/insights.md)

---

## Maintenance

### Daily
- Update [implementation-log.md](03-logs/implementation-log.md) when shipping significant code
- Log decisions in [decision-log.md](03-logs/decision-log.md) as they're made
- Track bugs in [bug-log.md](03-logs/bug-log.md)

### Weekly
- Review open questions in [assumptions.md](00-context/assumptions.md)
- Triage [bug-log.md](03-logs/bug-log.md) - resolve, prioritize, or document workarounds

### Monthly
- Update [system-map.md](00-context/system-map.md) with architecture changes
- Review [decision-log.md](03-logs/decision-log.md) - any decisions to revisit?
- Extract insights from [bug-log.md](03-logs/bug-log.md) to [insights.md](03-logs/insights.md)

### As Needed
- Update [prd.md](01-product/prd.md) when requirements change
- Update [vision.md](00-context/vision.md) if product direction shifts
- Add feature specs to `02-features/` when planning major features

---

## Documentation TODOs

- [ ] Add feature spec template to `02-features/feature-template/`
- [ ] Create `definition-of-done.md` with quality criteria
- [ ] Add `validation-log.md` to track post-ship reality vs expectations
- [ ] Create examples directory with working applications
- [ ] Add architecture diagrams (auto-generated from code if possible)
- [ ] Set up automated doc generation for API reference

---

## Questions?

- **For users**: Open GitHub issue with `question` label
- **For contributors**: See [dev-workflow.md](04-process/dev-workflow.md)
- **For AI assistants**: Load `.cursor/rules/` and ask specific questions

---

**The best documentation is documentation you actually maintain.**

This system is designed to be useful enough that you'll want to keep it updated. Start small, add as you go, let it grow with your project.
