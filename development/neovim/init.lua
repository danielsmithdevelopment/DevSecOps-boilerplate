-- Ensure packer is installed
local ensure_packer = function()
local fn = vim.fn
local install_path = fn.stdpath('data')..'/site/pack/packer/start/packer.nvim'
if fn.empty(fn.glob(install_path)) > 0 then
    fn.system({'git', 'clone', '--depth', '1', 'https://github.com/wbthomason/packer.nvim', install_path})
    vim.cmd [[packadd packer.nvim]]
    return true
end
return false
end

local packer_bootstrap = ensure_packer()
local packer = require('packer')
local keymap = vim.keymap

-- Plugin configurations
return packer.startup(function()
  use 'wbthomason/packer.nvim' -- Package manager

  -- Theme (Catpuccin-Mocha)
  use {
    "catppuccin/nvim",
    as = "catppuccin",
    config = function()
      vim.cmd("colorscheme catppuccin")
      vim.g.catppuccin_flavour = "mocha"
    end,
  }

  -- Telescope and dependencies
  use {
    'nvim-telescope/telescope.nvim',
    requires = {
        { 'nvim-lua/plenary.nvim' },
        { 'nvim-telescope/telescope-file-browser.nvim' },
        { 'nvim-telescope/telescope-symbols.nvim' },
        { 'nvim-telescope/telescope-git.nvim' },
    },
    config = function()
        -- Telescope setup and keymaps will be configured here after installation
        local telescope = require('telescope')
        telescope.setup({
            extensions = {
                file_browser = {},
                symbols = {},
                git = {}
            }
        })

        -- Key mappings
        keymap.set('n', '<space>e', ':Telescope file_browser<CR>', { desc = 'Explore files' })
        keymap.set('n', '<space>f', ':Telescope find_files<CR>', { desc = 'Find files' })
        keymap.set('n', '<space>g', ':Telescope git_commits<CR>', { desc = 'Search git commits' })
        keymap.set('n', '<space>s', ':Telescope symbols<CR>', { desc = 'Search symbols' })
    end
  }

  -- LSP
  use 'neovim/nvim-lspconfig'         -- LSP configuration
  use 'williamboman/mason.nvim'        -- Package manager for LSPs, formatters, etc.
  use 'williamboman/mason-lspconfig.nvim'
  use "jose-elias-alvarez/null-ls.nvim" -- For null-ls (additional LSP functionality)

  -- Completion
  use 'hrsh7th/nvim-cmp'               -- Completion plugin
  use 'hrsh7th/cmp-nvim-lsp'           -- LSP source for nvim-cmp
  use 'hrsh7th/cmp-buffer'             -- Buffer completion source
  use 'hrsh7th/cmp-path'              -- Path completion source

  -- UI Enhancements
  use 'nvim-treesitter/nvim-treesitter'-- Treesitter for syntax highlighting
  use 'kyazdani42/nvim-web-devicons'    -- File icons
  use {
    'nvim-lualine/lualine.nvim',
    requires = { 'nvim-web-devicons' },
    config = function()
      require('lualine').setup({
        options = {
          theme = 'catppuccin'
        }
      })
    end
  }

  -- File explorer
  use {
    "nvim-tree/nvim-tree.lua",
    requires = {
      "nvim-web-devicons", -- optional, for file icons
    },
    config = function()
      require("nvim-tree").setup({
        sort_by = "case_sensitive",
        view = {
          width = 30,
          mappings = {
            list = {
              { key = "<CR>", action = "edit" },
              { key = "v",     action = "vsplit" },
            },
          },
        },
      })
    end
  }

  -- Git
  use 'lewis6991/gitsigns.nvim'         -- Git signs plugin

  -- Snippets and templates
  use 'L3MON4D3/LuaSnip'                  -- Snippet engine
  use 'rafamadriz/friendly-snippets'       -- Collection of snippets

  -- Other useful plugins
  use 'ThePrimeagen/vim-behavior-plus'     -- Better behavior for Vim motions
  use 'numToStr/Comment.nvim'              -- Fast commenting plugin
  use {
    "windwp/nvim-autopairs",
    config = function()
      require("nvim-autopairs").setup()
    end
  }

  -- Go development
  use {
    'ray-x/go.nvim',
    requires = {
      'nvim-treesitter/nvim-treesitter',
      'neovim/nvim-lspconfig',
      'jose-elias-alvarez/null-ls.nvim'
    },
    config = function()
      require('go').setup({
        -- Add go.nvim configuration
        gofmt = 'gofumpt', -- Use gofumpt for better formatting
        max_line_len = 120,
        tag_transform = false,
        test_dir = '',
        comment_placeholder = '   ',
        lsp_cfg = true, -- false: use your own lspconfig
        lsp_gofumpt = true, -- true: set default gofmt in gopls
        lsp_on_attach = true, -- use on_attach from go.nvim
        dap_debug = true,
        lsp_inlay_hints = {
          enable = true,
        },
        lsp_keymaps = true,
        diagnostic = {
          hdlr = true,
          underline = true,
          virtual_text = true,
          signs = true,
        },
      })
    end
  }

  use {
    "olexsmir/gopher.nvim",
    ft = { "go", "gomod" }, -- only load on Go files
    requires = { -- Required dependencies
      "nvim-treesitter/nvim-treesitter", -- Treesitter required for gopher.nvim
      "neovim/nvim-lspconfig",
      "jose-elias-alvarez/null-ls.nvim"   -- For additional LS support
    },
  }

  use {
    'olexsmir/gotests-vim',
    ft = { "go", "gomod" }, -- only load on Go files
    requires = {
      "nvim-treesitter/nvim-treesitter",
    }
  }

  -- DAP for debugging (Go)
  use {
    "rcarriga/nvim-dap-Go",
    ft = { "go", "gomod" },
    dependencies = {
      'mfussenegger/nvim-dap',
      'nvim-telescope/telescope.nvim'
    }
  }

  -- Add basic LSP configuration
  use {
    'neovim/nvim-lspconfig',
    config = function()
      local lspconfig = require('lspconfig')
      -- Configure gopls for Go
      lspconfig.gopls.setup{
        capabilities = require('cmp_nvim_lsp').default_capabilities(),
        settings = {
          gopls = {
            analyses = {
              unusedparams = true,
            },
            staticcheck = true,
            gofumpt = true,
          },
        },
        on_attach = function(client, bufnr)
          -- Enable completion triggered by <c-x><c-o>
          vim.api.nvim_buf_set_option(bufnr, 'omnifunc', 'v:lua.vim.lsp.omnifunc')
          
          -- Format on save
          vim.api.nvim_create_autocmd("BufWritePre", {
            pattern = "*.go",
            callback = function()
              vim.lsp.buf.format()
            end,
          })
        end
      }
    end
  }

  -- Add Trouble for better error visualization
  use {
    "folke/trouble.nvim",
    requires = "kyazdani42/nvim-web-devicons",
    config = function()
      require("trouble").setup()
    end
  }

  -- Add which-key for better keymap discovery
  use {
    "folke/which-key.nvim",
    config = function()
      require("which-key").setup()
    end
  }

-- Key mappings
vim.keymap.set('n', '<space>e', ':Telescope file_browser<CR>', { desc = 'Explore files' })
vim.keymap.set('n', '<space>f', ':Telescope find_files<CR>', { desc = 'Find files' })
vim.keymap.set('n', '<space>g', ':Telescope git_commits<CR>', { desc = 'Search git commits' })
vim.keymap.set('n', '<space>s', ':Telescope symbols<CR>', { desc = 'Search symbols' })
vim.keymap.set('n', '<space>t', ':ToggleTerm<CR>', { desc = 'Toggle terminal' })
vim.keymap.set('n', '<space>v', ':vsplit<CR>', { desc = 'Vertical split' })

-- nvim-tree key mappings
vim.keymap.set('n', '<C-n>', ':NvimTreeToggle<CR>', { desc = 'Toggle file explorer' })
vim.keymap.set('n', '<C-p>', ':NvimTreeFindFile<CR>', { desc = 'Focus on current file in explorer' })

-- Git signs
vim.keymap.set('n', '[g', ':Gitsigns prev_hunk<CR>', { desc = 'Previous git hunk' })
vim.keymap.set('n', ']g', ':Gitsigns next_hunk<CR>', { desc = 'Next git hunk' })

-- Comment
vim.keymap.set('n', '<C-c>', '<Plug>CommentToggle<CR>', { desc = 'Toggle comment' })

-- LSP keymaps (corrected)
vim.keymap.set('n', 'gD', vim.lsp.buf.declaration, { desc = 'Go to declaration' })
vim.keymap.set('n', 'gd', vim.lsp.buf.definition, { desc = 'Go to definition' })
vim.keymap.set('n', 'gi', vim.lsp.buf.implementation, { desc = 'Go to implementation' })
vim.keymap.set('n', 'gr', vim.lsp.buf.references, { desc = 'Show references' })
vim.keymap.set('n', 'K', vim.lsp.buf.hover, { desc = 'Show hover documentation' })

-- Go-specific keymaps (enabled)
vim.keymap.set('n', '<space>go', ':GoRun<CR>', { desc = 'Run go program' })
vim.keymap.set('n', '<space>gb', ':GoBuild<CR>', { desc = 'Build go program' })
vim.keymap.set('n', '<space>gt', ':GoTest<CR>', { desc = 'Test go program' })
vim.keymap.set('n', '<space>gi', ':GoImport<CR>', { desc = 'Manage imports' })
vim.keymap.set('n', '<space>gf', ':GoFillStruct<CR>', { desc = 'Fill struct' })

-- DAP keymaps (for debugging)
vim.keymap.set('n', '<F5>', ':lua require"dap".continue()<CR>', { desc = 'Continue execution' })
vim.keymap.set('n', '<F1>', ':lua require"dap".toggle_breakpoint()<CR>', { desc = 'Toggle breakpoint' })

-- Lualine setup (already handled in config above)

-- Add to your completion setup
use {
  'hrsh7th/nvim-cmp',
  config = function()
    local cmp = require('cmp')
    cmp.setup({
      sources = {
        { name = 'nvim_lsp' },
        { name = 'luasnip' },
        { name = 'buffer' },
        { name = 'path' },
      },
      -- Add snippet support
      snippet = {
        expand = function(args)
          require('luasnip').lsp_expand(args.body)
        end,
      },
    })
  end
}

-- Add these keymaps for enhanced Go development
vim.keymap.set('n', '<space>gta', ':GoAddTag<CR>', { desc = 'Add tags to struct' })
vim.keymap.set('n', '<space>gtr', ':GoRmTag<CR>', { desc = 'Remove tags from struct' })
vim.keymap.set('n', '<space>gil', ':GoImplements<CR>', { desc = 'Show interfaces implemented' })
vim.keymap.set('n', '<space>gcv', ':GoCoverage<CR>', { desc = 'Show test coverage' })

if packer_bootstrap then
    packer.sync()
end
end)
