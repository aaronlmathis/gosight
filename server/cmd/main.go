/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// File: server/cmd/main.go
package main

import (
	"github.com/aaronlmathis/gosight/server/internal/bootstrap"
	"github.com/aaronlmathis/gosight/server/internal/server"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func main() {

	// Bootstrap config loading (flags -> env -> file)
	cfg := bootstrap.LoadServerConfig()

	// Initialize logging
	bootstrap.SetupLogging(cfg)

	grpcServer, listener, err := server.NewGRPCServer(cfg)
	if err != nil {
		utils.Fatal("Failed to start gRPC server: %v", err)
	}
	utils.Info("🚀 GoSight server listening on %s", cfg.ListenAddr)
	if err := grpcServer.Serve(listener); err != nil {
		utils.Fatal("Failed to serve: %v", err)
	}

}