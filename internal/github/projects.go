package github

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ProjectItem struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Number      int               `json:"number"`
	URL         string            `json:"url"`
	State       string            `json:"state"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Assignees   []string          `json:"assignees"`
	Labels      []string          `json:"labels"`
	Repository  string            `json:"repository"`
	Status      string            `json:"status"`
	ContentType string            `json:"contentType"`
	Fields      map[string]string `json:"fields,omitempty"`
}

type Project struct {
	Title string                 `json:"title"`
	Items map[string][]ProjectItem `json:"items"`
}

const projectItemsQueryUser = `query($owner: String!, $number: Int!, $cursor: String) {
  user(login: $owner) {
    projectV2(number: $number) {
      title
      items(first: 100, after: $cursor) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          id
          content {
            __typename
            ... on Issue {
              title
              number
              url
              state
              createdAt
              updatedAt
              assignees(first: 5) { nodes { login } }
              labels(first: 10) { nodes { name } }
              repository { nameWithOwner }
            }
            ... on PullRequest {
              title
              number
              url
              state
              createdAt
              updatedAt
              assignees(first: 5) { nodes { login } }
              labels(first: 10) { nodes { name } }
              repository { nameWithOwner }
            }
          }
          fieldValues(first: 20) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field { ... on ProjectV2SingleSelectField { name } }
              }
            }
          }
        }
      }
    }
  }
}`

const projectItemsQueryOrg = `query($owner: String!, $number: Int!, $cursor: String) {
  organization(login: $owner) {
    projectV2(number: $number) {
      title
      items(first: 100, after: $cursor) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          id
          content {
            __typename
            ... on Issue {
              title
              number
              url
              state
              createdAt
              updatedAt
              assignees(first: 5) { nodes { login } }
              labels(first: 10) { nodes { name } }
              repository { nameWithOwner }
            }
            ... on PullRequest {
              title
              number
              url
              state
              createdAt
              updatedAt
              assignees(first: 5) { nodes { login } }
              labels(first: 10) { nodes { name } }
              repository { nameWithOwner }
            }
          }
          fieldValues(first: 20) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field { ... on ProjectV2SingleSelectField { name } }
              }
            }
          }
        }
      }
    }
  }
}`

type projectV2Items struct {
	Title string `json:"title"`
	Items struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []struct {
			ID      string `json:"id"`
			Content *struct {
				Typename string    `json:"__typename"`
				Title    string    `json:"title"`
				Number   int       `json:"number"`
				URL      string    `json:"url"`
				State    string    `json:"state"`
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
				Assignees struct {
					Nodes []struct {
						Login string `json:"login"`
					} `json:"nodes"`
				} `json:"assignees"`
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"labels"`
				Repository struct {
					NameWithOwner string `json:"nameWithOwner"`
				} `json:"repository"`
			} `json:"content"`
			FieldValues struct {
				Nodes []struct {
					Name  string `json:"name"`
					Field struct {
						Name string `json:"name"`
					} `json:"field"`
				} `json:"nodes"`
			} `json:"fieldValues"`
		} `json:"nodes"`
	} `json:"items"`
}

type projectItemsResponse struct {
	Data struct {
		User struct {
			ProjectV2 *projectV2Items `json:"projectV2"`
		} `json:"user"`
		Organization struct {
			ProjectV2 *projectV2Items `json:"projectV2"`
		} `json:"organization"`
	} `json:"data"`
}

func (r *projectItemsResponse) projectV2() *projectV2Items {
	if r.Data.User.ProjectV2 != nil {
		return r.Data.User.ProjectV2
	}
	return r.Data.Organization.ProjectV2
}

const projectCacheTTL = 2 * time.Minute

type cachedProject struct {
	Project   *Project  `json:"project"`
	FetchedAt time.Time `json:"fetchedAt"`
}

func projectCachePath(owner string, number int) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("%s-%d", owner, number)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))[:12]
	return filepath.Join(base, "gh-planning", "projects", hash+".json"), nil
}

func loadProjectCache(owner string, number int) (*Project, bool) {
	path, err := projectCachePath(owner, number)
	if err != nil {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var cached cachedProject
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}
	if time.Since(cached.FetchedAt) > projectCacheTTL {
		return nil, false
	}
	return cached.Project, true
}

func saveProjectCache(owner string, number int, project *Project) {
	path, err := projectCachePath(owner, number)
	if err != nil {
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	data, err := json.Marshal(cachedProject{Project: project, FetchedAt: time.Now()})
	if err != nil {
		return
	}
	// Atomic write: temp file + rename to avoid partial writes.
	tmp, err := os.CreateTemp(dir, ".cache-*")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, writeErr := tmp.Write(data); writeErr != nil {
		tmp.Close()
		return
	}
	tmp.Close()
	_ = os.Rename(tmpName, path)
}

func GetProject(ctx context.Context, owner string, number int) (*Project, error) {
	if cached, ok := loadProjectCache(owner, number); ok {
		return cached, nil
	}

	fmt.Fprintf(os.Stderr, "Fetching project data...")

	project := &Project{Items: map[string][]ProjectItem{}}
	var cursor string
	page := 0
	// Try user query first; if the project isn't found, retry with org query.
	query := projectItemsQueryUser
	useOrg := false

	for {
		page++
		if page > 1 {
			fmt.Fprintf(os.Stderr, ".")
		}
		vars := map[string]interface{}{"owner": owner, "number": number}
		if cursor != "" {
			vars["cursor"] = cursor
		}
		payload, err := GraphQL(ctx, query, vars)
		if err != nil {
			if !useOrg {
				// User query failed, try org query from scratch
				useOrg = true
				query = projectItemsQueryOrg
				cursor = ""
				page = 0
				continue
			}
			return nil, err
		}
		var resp projectItemsResponse
		if err := json.Unmarshal(payload, &resp); err != nil {
			return nil, err
		}
		pv2 := resp.projectV2()
		if pv2 == nil {
			if !useOrg {
				// User query returned nil project, try org query
				useOrg = true
				query = projectItemsQueryOrg
				cursor = ""
				page = 0
				continue
			}
			return nil, errors.New("project not found")
		}
		if project.Title == "" {
			project.Title = pv2.Title
		}
		for _, node := range pv2.Items.Nodes {
			if node.Content == nil {
				continue
			}
			status := "No Status"
			fields := map[string]string{}
			for _, field := range node.FieldValues.Nodes {
				if field.Field.Name != "" && field.Name != "" {
					fields[field.Field.Name] = field.Name
				}
				if strings.EqualFold(field.Field.Name, "Status") && field.Name != "" {
					status = field.Name
				}
			}
			item := ProjectItem{
				ID:          node.ID,
				Title:       node.Content.Title,
				Number:      node.Content.Number,
				URL:         node.Content.URL,
				State:       node.Content.State,
				CreatedAt:   node.Content.CreatedAt,
				UpdatedAt:   node.Content.UpdatedAt,
				Repository:  node.Content.Repository.NameWithOwner,
				Status:      status,
				ContentType: node.Content.Typename,
				Fields:      fields,
			}
			for _, assignee := range node.Content.Assignees.Nodes {
				item.Assignees = append(item.Assignees, assignee.Login)
			}
			for _, label := range node.Content.Labels.Nodes {
				item.Labels = append(item.Labels, label.Name)
			}
			project.Items[status] = append(project.Items[status], item)
		}

		if !pv2.Items.PageInfo.HasNextPage {
			break
		}
		cursor = pv2.Items.PageInfo.EndCursor
	}

	fmt.Fprintf(os.Stderr, " done (%d items, cached for %s)\n", totalItems(project), projectCacheTTL)
	saveProjectCache(owner, number, project)
	return project, nil
}
const projectInfoQueryUser = `query($owner: String!, $number: Int!) {
  user(login: $owner) {
    projectV2(number: $number) {
      id
      title
      fields(first: 50) {
        nodes {
          ... on ProjectV2SingleSelectField {
            id
            name
            options { id name }
          }
        }
      }
    }
  }
}`

const projectInfoQueryOrg = `query($owner: String!, $number: Int!) {
  organization(login: $owner) {
    projectV2(number: $number) {
      id
      title
      fields(first: 50) {
        nodes {
          ... on ProjectV2SingleSelectField {
            id
            name
            options { id name }
          }
        }
      }
    }
  }
}`

type projectV2Info struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Fields struct {
		Nodes []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Options []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"options"`
		} `json:"nodes"`
	} `json:"fields"`
}

type projectInfoResponse struct {
	Data struct {
		User struct {
			ProjectV2 *projectV2Info `json:"projectV2"`
		} `json:"user"`
		Organization struct {
			ProjectV2 *projectV2Info `json:"projectV2"`
		} `json:"organization"`
	} `json:"data"`
}

func (r *projectInfoResponse) projectV2() *projectV2Info {
	if r.Data.User.ProjectV2 != nil {
		return r.Data.User.ProjectV2
	}
	return r.Data.Organization.ProjectV2
}

func GetProjectInfo(ctx context.Context, owner string, number int) (projectID string, title string, statusFieldID string, statusOptions map[string]string, err error) {
	vars := map[string]interface{}{"owner": owner, "number": number}

	// Try user query first
	payload, err := GraphQL(ctx, projectInfoQueryUser, vars)
	if err != nil {
		// User query failed, try org query
		payload, err = GraphQL(ctx, projectInfoQueryOrg, vars)
		if err != nil {
			return "", "", "", nil, err
		}
	}
	var resp projectInfoResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", "", "", nil, err
	}
	pv2 := resp.projectV2()
	if pv2 == nil {
		// User query returned nil, try org query
		payload, err = GraphQL(ctx, projectInfoQueryOrg, vars)
		if err != nil {
			return "", "", "", nil, err
		}
		var orgResp projectInfoResponse
		if err := json.Unmarshal(payload, &orgResp); err != nil {
			return "", "", "", nil, err
		}
		pv2 = orgResp.projectV2()
		if pv2 == nil {
			return "", "", "", nil, errors.New("project not found")
		}
	}
	statusOptions = map[string]string{}
	for _, field := range pv2.Fields.Nodes {
		if strings.EqualFold(field.Name, "Status") {
			statusFieldID = field.ID
			for _, option := range field.Options {
				statusOptions[option.Name] = option.ID
			}
			break
		}
	}
	return pv2.ID, pv2.Title, statusFieldID, statusOptions, nil
}

const addItemMutation = `mutation($projectId: ID!, $contentId: ID!) {
  addProjectV2ItemById(input: {projectId: $projectId, contentId: $contentId}) {
    item { id }
  }
}`

type addItemResponse struct {
	Data struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	} `json:"data"`
}

func AddItemToProject(ctx context.Context, projectID string, contentID string) (string, error) {
	payload, err := GraphQL(ctx, addItemMutation, map[string]interface{}{"projectId": projectID, "contentId": contentID})
	if err != nil {
		return "", err
	}
	var resp addItemResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", err
	}
	return resp.Data.AddProjectV2ItemById.Item.ID, nil
}

const updateStatusMutation = `mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
  updateProjectV2ItemFieldValue(input: {projectId: $projectId, itemId: $itemId, fieldId: $fieldId, value: {singleSelectOptionId: $optionId}}) {
    projectV2Item { id }
  }
}`

func UpdateItemStatus(ctx context.Context, projectID string, itemID string, fieldID string, optionID string) error {
	_, err := GraphQL(ctx, updateStatusMutation, map[string]interface{}{"projectId": projectID, "itemId": itemID, "fieldId": fieldID, "optionId": optionID})
	if err != nil {
		return err
	}
	return nil
}

// GetProjectField looks up a single-select field by name and returns its ID and options.
func GetProjectField(ctx context.Context, owner string, number int, fieldName string) (fieldID string, options map[string]string, err error) {
	vars := map[string]interface{}{"owner": owner, "number": number}

	payload, err := GraphQL(ctx, projectInfoQueryUser, vars)
	if err != nil {
		payload, err = GraphQL(ctx, projectInfoQueryOrg, vars)
		if err != nil {
			return "", nil, err
		}
	}
	var resp projectInfoResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", nil, err
	}
	pv2 := resp.projectV2()
	if pv2 == nil {
		payload, err = GraphQL(ctx, projectInfoQueryOrg, vars)
		if err != nil {
			return "", nil, err
		}
		var orgResp projectInfoResponse
		if err := json.Unmarshal(payload, &orgResp); err != nil {
			return "", nil, err
		}
		pv2 = orgResp.projectV2()
		if pv2 == nil {
			return "", nil, nil
		}
	}
	options = map[string]string{}
	for _, field := range pv2.Fields.Nodes {
		if strings.EqualFold(field.Name, fieldName) {
			fieldID = field.ID
			for _, option := range field.Options {
				options[option.Name] = option.ID
			}
			return fieldID, options, nil
		}
	}
	return "", nil, nil
}

func VerifyProject(ctx context.Context, owner string, number int) (string, error) {
	projectID, title, _, _, err := GetProjectInfo(ctx, owner, number)
	if err != nil {
		return "", err
	}
	if projectID == "" {
		return "", errors.New("project not found")
	}
	return title, nil
}

func totalItems(p *Project) int {
count := 0
for _, items := range p.Items {
count += len(items)
}
return count
}
