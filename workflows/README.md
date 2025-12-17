# n8n Workflows

This directory contains n8n workflow definitions for the Cyber News project.

## Workflows

### cyber-news-aggregator.json
Main workflow for aggregating cyber security news from multiple sources, enriching content with AI analysis, and storing results in the database. Includes:
- News source integration
- Content enrichment via AI
- Data persistence
- Error handling and notifications

## Usage

Import these workflows into your n8n instance via the n8n UI using **Import from file** or programmatically via the n8n API.

## Directory Structure

```
workflows/
├── README.md                    # This file
└── cyber-news-aggregator.json   # Main aggregation workflow
```

## Notes

- Keep workflows organized by function or feature
- Update this README when adding new workflows
- Store sensitive credentials in n8n environment variables, not in workflow JSON
