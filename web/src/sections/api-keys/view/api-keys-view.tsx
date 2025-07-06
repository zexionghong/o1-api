import { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Table from '@mui/material/Table';
import Button from '@mui/material/Button';
import TableBody from '@mui/material/TableBody';
import Typography from '@mui/material/Typography';
import TableContainer from '@mui/material/TableContainer';
import TablePagination from '@mui/material/TablePagination';
import Alert from '@mui/material/Alert';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';

import { Iconify } from 'src/components/iconify';
import { useAuth } from 'src/contexts/auth-context';
import { DashboardContent } from 'src/layouts/dashboard';
import { Scrollbar } from 'src/components/scrollbar';

import { ApiKeyTableRow } from '../api-key-table-row';
import { ApiKeyTableHead } from '../api-key-table-head';
import { ApiKeyTableToolbar } from '../api-key-table-toolbar';
import { ApiKeyDetailDialog } from '../api-key-detail-dialog';
import { ApiKeyCreatedDialog } from '../api-key-created-dialog';

// ----------------------------------------------------------------------

interface ApiKey {
  id: number;
  name: string;
  key?: string; // 完整的API密钥
  key_prefix: string;
  status: 'active' | 'inactive' | 'revoked';
  permissions?: {
    allowed_providers?: string[];
    allowed_models?: string[];
  };
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

interface CreateApiKeyData {
  name?: string;
}

// ----------------------------------------------------------------------

export function ApiKeysView() {
  const { state } = useAuth();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const [filterName, setFilterName] = useState('');
  const [selected, setSelected] = useState<number[]>([]);

  // Dialog states
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [openDetailDialog, setOpenDetailDialog] = useState(false);
  const [openCreatedDialog, setOpenCreatedDialog] = useState(false);
  const [selectedApiKey, setSelectedApiKey] = useState<ApiKey | null>(null);
  const [newApiKey, setNewApiKey] = useState<{ key: string; name: string } | null>(null);
  const [createFormData, setCreateFormData] = useState<CreateApiKeyData>({
    name: '',
  });

  // 获取API密钥列表
  const fetchApiKeys = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('No access token found');
      }

      const response = await fetch(`http://localhost:8080/admin/users/${state.user?.id}/api-keys`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch API keys');
      }

      const result = await response.json();
      setApiKeys(result.data || []);
    } catch (err) {
      console.error('Error fetching API keys:', err);
      setError(err instanceof Error ? err.message : 'Failed to load API keys');
    } finally {
      setLoading(false);
    }
  };

  // 生成随机名称
  const generateRandomName = () => {
    const adjectives = ['Swift', 'Bright', 'Smart', 'Quick', 'Fast', 'Cool', 'Sharp', 'Bold'];
    const nouns = ['Key', 'Token', 'Access', 'Gate', 'Bridge', 'Link', 'Path', 'Code'];
    const randomAdjective = adjectives[Math.floor(Math.random() * adjectives.length)];
    const randomNoun = nouns[Math.floor(Math.random() * nouns.length)];
    const randomNumber = Math.floor(Math.random() * 1000);
    return `${randomAdjective}${randomNoun}${randomNumber}`;
  };

  // 创建API密钥
  const handleCreateApiKey = async () => {
    try {
      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('No access token found');
      }

      // 准备请求数据
      const requestData: any = {
        user_id: state.user?.id,
      };

      // 只有当用户输入了名称时才发送name字段
      if (createFormData.name?.trim()) {
        requestData.name = createFormData.name.trim();
      }

      const response = await fetch('http://localhost:8080/admin/api-keys/', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
      });

      if (!response.ok) {
        throw new Error('Failed to create API key');
      }

      const result = await response.json();

      // 显示创建成功对话框，包含完整的API Key
      if (result.success && result.data) {
        setNewApiKey({
          key: result.data.key,
          name: result.data.name || 'Unnamed API Key',
        });
        setOpenCreatedDialog(true);
      }

      setOpenCreateDialog(false);
      setCreateFormData({
        name: '',
      });
      fetchApiKeys();
    } catch (err) {
      console.error('Error creating API key:', err);
      setError(err instanceof Error ? err.message : 'Failed to create API key');
    }
  };

  // 更改API密钥状态
  const handleStatusChange = async (id: number, status: 'active' | 'inactive') => {
    try {
      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('No access token found');
      }

      const response = await fetch(`http://localhost:8080/admin/api-keys/${id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: status,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to update API key status');
      }

      fetchApiKeys();
      // 如果详情对话框打开，更新选中的API密钥
      if (selectedApiKey && selectedApiKey.id === id) {
        setSelectedApiKey({ ...selectedApiKey, status });
      }
    } catch (err) {
      console.error('Error updating API key status:', err);
      setError(err instanceof Error ? err.message : 'Failed to update API key status');
    }
  };

  // 软删除API密钥
  const handleDeleteApiKey = async (id: number) => {
    try {
      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('No access token found');
      }

      // 使用软删除，将状态设置为revoked
      const response = await fetch(`http://localhost:8080/admin/api-keys/${id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: 'revoked',
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to delete API key');
      }

      fetchApiKeys();
      // 关闭详情对话框
      setOpenDetailDialog(false);
      setSelectedApiKey(null);
    } catch (err) {
      console.error('Error deleting API key:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete API key');
    }
  };

  // 查看API密钥详情
  const handleViewDetails = (apiKey: ApiKey) => {
    setSelectedApiKey(apiKey);
    setOpenDetailDialog(true);
  };

  useEffect(() => {
    if (state.user?.id) {
      fetchApiKeys();
    }
  }, [state.user?.id]);

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelecteds = apiKeys.map((n) => n.id);
      setSelected(newSelecteds);
      return;
    }
    setSelected([]);
  };

  const handleClick = (event: React.MouseEvent<unknown>, id: number) => {
    const selectedIndex = selected.indexOf(id);
    let newSelected: number[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selected, id);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selected.slice(1));
    } else if (selectedIndex === selected.length - 1) {
      newSelected = newSelected.concat(selected.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selected.slice(0, selectedIndex),
        selected.slice(selectedIndex + 1)
      );
    }
    setSelected(newSelected);
  };

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setPage(0);
    setRowsPerPage(parseInt(event.target.value, 10));
  };

  const handleFilterByName = (event: React.ChangeEvent<HTMLInputElement>) => {
    setPage(0);
    setFilterName(event.target.value);
  };

  const filteredApiKeys = apiKeys.filter((apiKey) =>
    apiKey.name.toLowerCase().indexOf(filterName.toLowerCase()) !== -1
  );

  const isSelected = (id: number) => selected.indexOf(id) !== -1;

  const emptyRows = page > 0 ? Math.max(0, (1 + page) * rowsPerPage - filteredApiKeys.length) : 0;

  if (loading) {
    return (
      <DashboardContent>
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px' }}>
          <CircularProgress />
        </Box>
      </DashboardContent>
    );
  }

  return (
    <DashboardContent>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 5 }}>
        <Typography variant="h4" sx={{ flexGrow: 1 }}>
          API Keys
        </Typography>
        <Button
          variant="contained"
          color="inherit"
          startIcon={<Iconify icon="solar:pen-bold" />}
          onClick={() => setOpenCreateDialog(true)}
        >
          New API Key
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Card>
        <ApiKeyTableToolbar
          numSelected={selected.length}
          filterName={filterName}
          onFilterName={handleFilterByName}
        />

        <Scrollbar>
          <TableContainer sx={{ overflow: 'unset' }}>
            <Table sx={{ minWidth: 800 }}>
              <ApiKeyTableHead
                order="asc"
                orderBy="name"
                rowCount={filteredApiKeys.length}
                numSelected={selected.length}
                onRequestSort={() => {}}
                onSelectAllClick={handleSelectAllClick}
                headLabel={[
                  { id: 'name', label: 'Name' },
                  { id: 'key_prefix', label: 'Key Prefix' },
                  { id: 'status', label: 'Status' },
                  { id: 'last_used_at', label: 'Last Used' },
                  { id: 'created_at', label: 'Created' },
                  { id: '' },
                ]}
              />
              <TableBody>
                {filteredApiKeys
                  .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                  .map((row) => (
                    <ApiKeyTableRow
                      key={row.id}
                      row={row}
                      selected={isSelected(row.id)}
                      onSelectRow={(event) => handleClick(event, row.id)}
                      onViewDetails={() => handleViewDetails(row)}
                      onStatusChange={(status) => handleStatusChange(row.id, status)}
                      onDeleteRow={() => handleDeleteApiKey(row.id)}
                    />
                  ))}
                {emptyRows > 0 && (
                  <tr style={{ height: 53 * emptyRows }}>
                    <td colSpan={6} />
                  </tr>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </Scrollbar>

        <TablePagination
          page={page}
          component="div"
          count={filteredApiKeys.length}
          rowsPerPage={rowsPerPage}
          onPageChange={handleChangePage}
          rowsPerPageOptions={[5, 10, 25]}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Card>

      {/* Create API Key Dialog */}
      <Dialog open={openCreateDialog} onClose={() => setOpenCreateDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New API Key</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <TextField
              fullWidth
              label="API Key Name (Optional)"
              placeholder="Leave empty to auto-generate"
              value={createFormData.name}
              onChange={(e) => setCreateFormData({ ...createFormData, name: e.target.value })}
              helperText="If left empty, a random name will be generated automatically"
              sx={{ mb: 2 }}
            />
            <Button
              variant="text"
              size="small"
              onClick={() => setCreateFormData({ ...createFormData, name: generateRandomName() })}
              sx={{ mb: 3 }}
            >
              Generate Random Name
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenCreateDialog(false)}>Cancel</Button>
          <Button onClick={handleCreateApiKey} variant="contained">
            Create
          </Button>
        </DialogActions>
      </Dialog>

      {/* API Key Detail Dialog */}
      <ApiKeyDetailDialog
        open={openDetailDialog}
        apiKey={selectedApiKey}
        onClose={() => {
          setOpenDetailDialog(false);
          setSelectedApiKey(null);
        }}
        onStatusChange={handleStatusChange}
        onDelete={handleDeleteApiKey}
      />

      {/* API Key Created Dialog */}
      <ApiKeyCreatedDialog
        open={openCreatedDialog}
        apiKey={newApiKey?.key || ''}
        apiKeyName={newApiKey?.name || ''}
        onClose={() => {
          setOpenCreatedDialog(false);
          setNewApiKey(null);
        }}
      />
    </DashboardContent>
  );
}
