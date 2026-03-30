# AI Assistant

Tookly includes a workflow-oriented assistant that helps teams query data, draft content, structure documentation, and execute operations — all within the platform.

The assistant is provider-agnostic. Administrators configure the AI provider per workspace or instance. Without a configured provider, AI features are unavailable but Tookly works normally.

---

## Provider configuration

The assistant connects to any LLM provider through a unified configuration:

- **Endpoint URL** — the provider's API endpoint.
- **API key** — authentication credential for the provider.
- **Model name** — the specific model to use.

Supported providers include OpenAI (GPT), Anthropic (Claude), Google (Gemini), Ollama, vLLM, and any API-compatible service. Configuration is set per workspace or per instance depending on the deployment.

When no provider is configured, AI features are not available — the rest of Tookly remains fully functional.

---

## Assistant capabilities

The assistant operates as a copilot for workflow, not as an autonomous agent.

**Query**
- Search and retrieve data from the workspace: issues, boards, projects, members, statuses.
- Answer questions about project state, progress, and history.

**Draft and structure**
- Assist in writing documentation pages, decision records, and project descriptions.
- Suggest structure for new pages using predefined templates.
- Help refine issue titles, descriptions, and acceptance criteria.

**Execute**
- Create, update, move, or archive issues, statuses, boards, and other entities.
- All operations go through the normal Tookly API under the authenticated user's session.

---

## Execution model

The assistant acts **as the user**. Every operation it performs uses the authenticated user's session and permissions.

- If the user is a workspace member, the assistant can read and work on issues.
- If the user is a workspace admin, the assistant can also manage members, create projects, and configure workflows.
- If the user lacks permissions for an action, the assistant's request fails with the same error as if performed through the UI or API directly.

There is no AI superuser. The assistant cannot bypass authorization, access other workspaces, or perform actions the user could not do themselves.

---

## Proposals

A `Proposal` represents a suggested change to an entity in Tookly. Proposals can originate from a human or from the AI assistant.

| Field | Description |
|---|---|
| `origin` | `human` or `ai` |
| `author` | The authenticated user who created or triggered the proposal |
| `target` | The entity being changed (document, issue, backlog item, etc.) |
| `payload` | The proposed action or content |

Both human and AI proposals follow the same execution flow and the same permission model. The origin field distinguishes who initiated the proposal, but the system treats them identically for authorization and execution.

---

## Documentation

Project documentation pages are persisted as editable Markdown.

**Templates and forms**
- Predefined templates help structure initial content for common page types: decision records, architecture documents, runbooks, meeting notes.
- Templates provide a starting structure — after creation, the page is free-form Markdown that can be edited without constraints.

Documentation in Tookly is stored as editable Markdown content within each project. Templates help structure the initial content, but pages remain flexible and user-editable. The assistant can help draft and structure content without imposing constraints on the final result.

---

## MCP integration

Tookly supports the Model Context Protocol (MCP) for connecting to external systems.

**Capabilities**
- Read data from external sources: Git repositories, CI pipelines, messaging platforms, monitoring systems.
- Execute actions in external systems when supported by the connector.
- Provide external context to the assistant for more informed responses.

**Authorization**
- MCP connectors operate within the scope of the authenticated user's session and permissions.

**Extensibility**
- Self-hosted environments can add custom MCP connectors for internal systems.
- The connector model is designed to be extensible without modifying Tookly core.
