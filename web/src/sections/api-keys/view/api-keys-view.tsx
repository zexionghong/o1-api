import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Table from '@mui/material/Table';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import TableBody from '@mui/material/TableBody';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TableContainer from '@mui/material/TableContainer';
import TablePagination from '@mui/material/TablePagination';
import CircularProgress from '@mui/material/CircularProgress';

import { useAuth } from 'src/contexts/auth-context';
import { DashboardContent } from 'src/layouts/dashboard';
import api from 'src/services/api';

import { Iconify } from 'src/components/iconify';
import { Scrollbar } from 'src/components/scrollbar';

import { ApiKeyTableRow } from '../api-key-table-row';
import { ApiKeyTableHead } from '../api-key-table-head';
import { ApiKeyTableToolbar } from '../api-key-table-toolbar';
import { ApiKeyDetailDialog } from '../api-key-detail-dialog';
import { ApiKeyCreatedDialog } from '../api-key-created-dialog';
import { QuotaConfigDialog } from '../quota-config-dialog';

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
  const { t } = useTranslation();
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
  const [openQuotaDialog, setOpenQuotaDialog] = useState(false);
  const [selectedApiKey, setSelectedApiKey] = useState<ApiKey | null>(null);
  const [newApiKey, setNewApiKey] = useState<{ key: string; name: string } | null>(null);
  const [createFormData, setCreateFormData] = useState<CreateApiKeyData>({
    name: '',
  });

  // 获取API密钥列表
  const fetchApiKeys = useCallback(async () => {
    try {
      setLoading(true);
      if (!state.user?.id) {
        throw new Error('User ID not found');
      }

      const response = await api.get(`/admin/users/${state.user.id}/api-keys`);
      setApiKeys(response.data || []);
    } catch (err) {
      console.error('Error fetching API keys:', err);
      setError(err instanceof Error ? err.message : 'Failed to load API keys');
    } finally {
      setLoading(false);
    }
  }, [state.user?.id]);

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

      const response = await api.post('/admin/api-keys/', requestData);

      if (!response.success) {
        throw new Error('Failed to create API key');
      }

      // 显示创建成功对话框，包含完整的API Key
      if (response.success && response.data) {
        setNewApiKey({
          key: response.data.key,
          name: response.data.name || 'Unnamed API Key',
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

      const response = await api.put(`/admin/api-keys/${id}`, { status });

      if (!response.success) {
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
      const response = await api.put(`/admin/api-keys/${id}`, { status: 'revoked' });

      if (!response.success) {
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

  // 配置API密钥配额
  const handleConfigureQuota = (apiKey: ApiKey) => {
    setSelectedApiKey(apiKey);
    setOpenQuotaDialog(true);
  };

  useEffect(() => {
    if (state.user?.id) {
      fetchApiKeys();
    }
  }, [state.user?.id, fetchApiKeys]);

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
          {t('api_keys.title')}
        </Typography>
        <Button
          variant="contained"
          color="inherit"
          startIcon={<Iconify icon="solar:pen-bold" />}
          onClick={() => setOpenCreateDialog(true)}
        >
          {t('api_keys.create_key')}
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
                  { id: 'name', label: t('common.name') },
                  { id: 'key_prefix', label: t('api_keys.key_prefix') },
                  { id: 'status', label: t('common.status') },
                  { id: 'quotas', label: '限额配置', align: 'center' as const },
                  { id: 'last_used_at', label: t('api_keys.last_used') },
                  { id: 'created_at', label: t('common.created_at') },
                  { id: '', label: '' },
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
                      onConfigureQuota={() => handleConfigureQuota(row)}
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
        <DialogTitle>{t('api_keys.create_key')}</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <TextField
              fullWidth
              label={t('api_keys.key_name')}
              placeholder={t('api_keys.name_placeholder')}
              value={createFormData.name}
              onChange={(e) => setCreateFormData({ ...createFormData, name: e.target.value })}
              helperText={t('api_keys.name_helper_text')}
              sx={{ mb: 2 }}
            />
            <Button
              variant="text"
              size="small"
              onClick={() => setCreateFormData({ ...createFormData, name: generateRandomName() })}
              sx={{ mb: 3 }}
            >
              {t('api_keys.generate_random_name')}
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenCreateDialog(false)}>{t('common.cancel')}</Button>
          <Button onClick={handleCreateApiKey} variant="contained">
            {t('common.create')}
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

      {/* Quota Config Dialog */}
      {selectedApiKey && (
        <QuotaConfigDialog
          open={openQuotaDialog}
          onClose={() => {
            setOpenQuotaDialog(false);
            setSelectedApiKey(null);
          }}
          apiKeyId={selectedApiKey.id}
          apiKeyName={selectedApiKey.name}
        />
      )}
    </DashboardContent>
  );
}
