// Copyright 2019 The Protocol Buffers Language Server Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-language-server/protocol"
	"github.com/go-language-server/uri"
)

func (s *Server) initialize(ctx context.Context, params *protocol.InitializeParams) (result *protocol.InitializeResult, err error) {
	s.stateMu.RLock()
	state := s.state
	s.stateMu.RUnlock()
	if state > stateInitializing {
		err = errors.New("server already initialized")
		return
	}
	s.stateMu.Lock()
	s.state = stateInitializing
	s.stateMu.Unlock()

	folders := params.WorkspaceFolders
	if len(folders) == 0 {
		if rootURI := params.RootURI; rootURI != "" {
			folders = []protocol.WorkspaceFolder{
				{
					URI:  rootURI.Filename(),
					Name: filepath.Base(rootURI.Filename()),
				},
			}
		} else {
			err = errors.New("single file mode not supported yet")
			return
		}
	}

	for _, folder := range folders {
		s.session.NewView(ctx, folder.Name, uri.File(folder.URI))
	}

	result = &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: nil,
			HoverProvider:    false,
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"."},
			},
			SignatureHelpProvider: &protocol.SignatureHelpOptions{
				TriggerCharacters: nil,
			},
			DefinitionProvider:              false,
			WorkspaceSymbolProvider:         false,
			DocumentFormattingProvider:      false,
			DocumentRangeFormattingProvider: false,
			RenameProvider:                  nil,
			FoldingRangeProvider:            nil,
			Workspace: &protocol.ServerCapabilitiesWorkspace{
				WorkspaceFolders: &protocol.ServerCapabilitiesWorkspaceFolders{
					Supported:           false,
					ChangeNotifications: nil,
				},
			},
		},
	}
	return
}

func (s *Server) initialized(context.Context, *protocol.InitializedParams) (err error) {
	s.stateMu.Lock()
	s.state = stateInitialized
	s.stateMu.Unlock()
	return
}

func (s *Server) shutdown(context.Context) (err error) {
	s.stateMu.RLock()
	state := s.state
	s.stateMu.RUnlock()
	if state < stateInitialized {
		err = errors.New("server not initialized")
		return
	}
	s.stateMu.Lock()
	s.state = stateShutdown
	s.stateMu.Unlock()
	return
}

func (s *Server) exit(context.Context) (err error) {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	if s.state != stateShutdown {
		os.Exit(1)
	}
	os.Exit(0)
	return
}
