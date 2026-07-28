package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// File is a manifest bound to its on-disk YAML document.
// Saving updates the existing AST to preserve comments and key order.
type File struct {
	Manifest *Manifest
	path     string
	doc      *yaml.Node
}

// Open reads kusabi.yaml and retains its YAML document structure.
func Open(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.Repositories == nil {
		m.Repositories = make(map[string]Repository)
	}
	m.RepositoryOrder = extractRepositoryOrder(&doc)
	if err := Validate(&m); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}

	return &File{Manifest: &m, path: path, doc: &doc}, nil
}

// Path returns the manifest file path.
func (f *File) Path() string { return f.path }

// RootDir returns the workspace root containing kusabi.yaml.
func (f *File) RootDir() string { return filepath.Dir(f.path) }

// Save writes the manifest, preserving comments and repository key order.
func (f *File) Save() error {
	if err := Validate(f.Manifest); err != nil {
		return fmt.Errorf("validate manifest: %w", err)
	}
	syncToNode(f.Manifest, f.doc)

	data, err := yaml.Marshal(f.doc)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(f.path), ".kusabi-*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, f.path); err != nil {
		return fmt.Errorf("rename to %s: %w", f.path, err)
	}
	return nil
}

// Load reads kusabi.yaml. Prefer Open when the file will be saved back.
func Load(path string) (*Manifest, error) {
	f, err := Open(path)
	if err != nil {
		return nil, err
	}
	return f.Manifest, nil
}

// Save writes the manifest at path. Uses Open when the file exists to preserve YAML structure.
func Save(m *Manifest, path string) error {
	f, err := Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return saveNew(m, path)
		}
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist) {
			return saveNew(m, path)
		}
		return err
	}
	preservedOrder := f.Manifest.RepositoryOrder
	f.Manifest = m
	if len(f.Manifest.RepositoryOrder) == 0 && len(preservedOrder) > 0 {
		f.Manifest.RepositoryOrder = preservedOrder
	}
	return f.Save()
}

func saveNew(m *Manifest, path string) error {
	if err := Validate(m); err != nil {
		return fmt.Errorf("validate manifest: %w", err)
	}

	doc := buildDocumentNode(m)
	data, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".kusabi-*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
}

// RepositoryNames returns repository names in declaration order.
func (m *Manifest) RepositoryNames() []string {
	if len(m.RepositoryOrder) > 0 {
		seen := make(map[string]bool, len(m.Repositories))
		names := make([]string, 0, len(m.Repositories))
		for _, name := range m.RepositoryOrder {
			if _, ok := m.Repositories[name]; ok {
				names = append(names, name)
				seen[name] = true
			}
		}
		var extras []string
		for name := range m.Repositories {
			if !seen[name] {
				extras = append(extras, name)
			}
		}
		sort.Strings(extras)
		return append(names, extras...)
	}
	names := make([]string, 0, len(m.Repositories))
	for name := range m.Repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func extractRepositoryOrder(doc *yaml.Node) []string {
	root := documentRoot(doc)
	if root == nil {
		return nil
	}
	repos := mappingValue(root, "repositories")
	if repos == nil || repos.Kind != yaml.MappingNode {
		return nil
	}
	order := make([]string, 0, len(repos.Content)/2)
	for i := 0; i < len(repos.Content); i += 2 {
		order = append(order, repos.Content[i].Value)
	}
	return order
}

func syncToNode(m *Manifest, doc *yaml.Node) {
	root := documentRoot(doc)
	if root == nil {
		return
	}
	setScalar(root, "version", m.Version)
	setScalar(root, "name", m.Name)
	if m.Description != "" {
		setScalar(root, "description", m.Description)
	} else {
		removeKey(root, "description")
	}
	syncContextNode(m.Context, root)
	syncRepositoriesNode(m, root)
}

func syncContextNode(cfg ContextConfig, root *yaml.Node) {
	ctx := mappingValue(root, "context")
	if cfg.Agents == "" && len(cfg.Includes) == 0 {
		removeKey(root, "context")
		return
	}
	if ctx == nil {
		ctx = &yaml.Node{Kind: yaml.MappingNode}
		setMappingValue(root, "context", ctx)
	}
	if cfg.Agents != "" {
		setScalar(ctx, "agents", cfg.Agents)
	} else {
		removeKey(ctx, "agents")
	}
	syncSequence(ctx, "includes", cfg.Includes)
}

func syncRepositoriesNode(m *Manifest, root *yaml.Node) {
	repos := mappingValue(root, "repositories")
	if repos == nil {
		repos = &yaml.Node{Kind: yaml.MappingNode}
		setMappingValue(root, "repositories", repos)
	}

	existingKeys := map[string]*yaml.Node{}
	existingVals := map[string]*yaml.Node{}
	for i := 0; i < len(repos.Content); i += 2 {
		existingKeys[repos.Content[i].Value] = repos.Content[i]
		existingVals[repos.Content[i].Value] = repos.Content[i+1]
	}

	order := m.RepositoryNames()
	var content []*yaml.Node
	for _, name := range order {
		repo, ok := m.Repositories[name]
		if !ok {
			continue
		}
		key := existingKeys[name]
		if key == nil {
			key = &yaml.Node{Kind: yaml.ScalarNode, Value: name}
		}
		val := existingVals[name]
		if val == nil {
			val = &yaml.Node{Kind: yaml.MappingNode}
		}
		updateRepoNode(val, repo)
		content = append(content, key, val)
	}
	repos.Content = content
}

func updateRepoNode(node *yaml.Node, repo Repository) {
	setScalar(node, "path", repo.Path)
	setScalar(node, "url", repo.URL)
	if repo.Role != "" {
		setScalar(node, "role", repo.Role)
	} else {
		removeKey(node, "role")
	}
	if len(repo.Tags) > 0 {
		syncSequence(node, "tags", repo.Tags)
	} else {
		removeKey(node, "tags")
	}
}

func buildDocumentNode(m *Manifest) *yaml.Node {
	root := &yaml.Node{Kind: yaml.MappingNode}
	setScalar(root, "version", m.Version)
	setScalar(root, "name", m.Name)
	if m.Description != "" {
		setScalar(root, "description", m.Description)
	}
	if m.Context.Agents != "" || len(m.Context.Includes) > 0 {
		ctx := &yaml.Node{Kind: yaml.MappingNode}
		if m.Context.Agents != "" {
			setScalar(ctx, "agents", m.Context.Agents)
		}
		syncSequence(ctx, "includes", m.Context.Includes)
		setMappingValue(root, "context", ctx)
	}
	syncRepositoriesNode(m, root)
	return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
}

func documentRoot(doc *yaml.Node) *yaml.Node {
	if doc == nil || doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	return doc.Content[0]
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func setMappingValue(node *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1] = value
			return
		}
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		value,
	)
}

func setScalar(node *yaml.Node, key, value string) {
	existing := mappingValue(node, key)
	if existing != nil {
		existing.Value = value
		return
	}
	setMappingValue(node, key, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
}

func syncSequence(node *yaml.Node, key string, values []string) {
	seq := mappingValue(node, key)
	if seq == nil {
		seq = &yaml.Node{Kind: yaml.SequenceNode}
		setMappingValue(node, key, seq)
	}
	content := make([]*yaml.Node, 0, len(values))
	for i, v := range values {
		if i < len(seq.Content) {
			seq.Content[i].Value = v
			content = append(content, seq.Content[i])
		} else {
			content = append(content, &yaml.Node{Kind: yaml.ScalarNode, Value: v})
		}
	}
	seq.Content = content
}

func removeKey(node *yaml.Node, key string) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content = append(node.Content[:i], node.Content[i+2:]...)
			return
		}
	}
}
