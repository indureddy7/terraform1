---
layout: "azuredevops"
page_title: "AzureDevops: azuredevops_wiki"
description: |-
  Manages Wikis within Azure DevOps project.
---

# azuredevops_wiki

Manages Wikis within Azure DevOps project.

## Example Usage

```hcl
resource "azuredevops_project" "project" {
  name = "My Awesome Project"
  description  = "All of my awesomee things"
}

resource "azuredevops_git_repository" "repository" {
  project_id = azuredevops_project.project.id
  name       = "My Awesome Repo"
  initialization {
    init_type = "Clean"
  }
}

resource "azuredevops_wiki" "test" {
  name = "project wiki "
  project_id = azuredevops_project.project.id
  type = "projectWiki"
}

resource "azuredevops_wiki" "test2" {
  name = "additional  wiki in repo"
  project_id = azuredevops_project.project.id
  repository_id = azuredevops_git_repository.repository.id
  versions = "main"
  type = "codeWiki"
  mappedpath = "/"
}
```

## Argument Reference

The following arguments are supported:

- `project_id` - (Required) The project ID.
- `name` - (Required) The name of the Wiki.
- `type` -  (Required) The type of the wiki.
- `repository_id` - (Optional) The repository ID. Not required for ProjectWiki type.
- `version` - (Optional) Version of the wiki. Not required for ProjectWiki type.
- `mappedpath` - (Optional) Folder path inside repository which is shown as Wiki. Not required for ProjectWiki type.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the wiki returned after creation in Azure DevOps.

## Relevant Links

- [Azure DevOps Service REST API 6.0 - Wiki ](https://learn.microsoft.com/en-us/rest/api/azure/devops/wiki/wikis?view=azure-devops-rest-6.0)
