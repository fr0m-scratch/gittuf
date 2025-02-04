// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/gittuf/gittuf/internal/cmd/policy/addkey"
	"github.com/gittuf/gittuf/internal/cmd/policy/addperson"
	"github.com/gittuf/gittuf/internal/cmd/policy/addrule"
	i "github.com/gittuf/gittuf/internal/cmd/policy/init"
	"github.com/gittuf/gittuf/internal/cmd/policy/listprincipals"
	"github.com/gittuf/gittuf/internal/cmd/policy/listrules"
	"github.com/gittuf/gittuf/internal/cmd/policy/persistent"
	"github.com/gittuf/gittuf/internal/cmd/policy/removekey"
	"github.com/gittuf/gittuf/internal/cmd/policy/removeperson"
	"github.com/gittuf/gittuf/internal/cmd/policy/removerule"
	"github.com/gittuf/gittuf/internal/cmd/policy/reorderrules"
	"github.com/gittuf/gittuf/internal/cmd/policy/sign"
	tui "github.com/gittuf/gittuf/internal/cmd/policy/tui"
	"github.com/gittuf/gittuf/internal/cmd/policy/updaterule"
	"github.com/gittuf/gittuf/internal/cmd/trustpolicy/apply"
	"github.com/gittuf/gittuf/internal/cmd/trustpolicy/discard"
	"github.com/gittuf/gittuf/internal/cmd/trustpolicy/remote"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	o := &persistent.Options{}
	cmd := &cobra.Command{
		Use:               "policy",
		Short:             "Tools to manage gittuf policies",
		DisableAutoGenTag: true,
	}
	o.AddPersistentFlags(cmd)

	cmd.AddCommand(i.New(o))
	cmd.AddCommand(addkey.New(o))
	cmd.AddCommand(removekey.New(o))
	cmd.AddCommand(addperson.New(o))
	cmd.AddCommand(removeperson.New(o))
	cmd.AddCommand(apply.New())
	cmd.AddCommand(discard.New())
	cmd.AddCommand(addrule.New(o))
	cmd.AddCommand(listprincipals.New())
	cmd.AddCommand(listrules.New())
	cmd.AddCommand(remote.New())
	cmd.AddCommand(removerule.New(o))
	cmd.AddCommand(reorderrules.New(o))
	cmd.AddCommand(sign.New(o))
	cmd.AddCommand(updaterule.New(o))
	cmd.AddCommand(tui.New(o))

	return cmd
}
