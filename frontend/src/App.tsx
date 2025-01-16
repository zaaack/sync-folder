import { useEffect, useState } from 'react'
import './App.css'
import {
  Button,
  Input,
  notification,
  Space,
  Table,
  type TableColumnProps,
} from 'antd'
import {
  Greet,
  LoadConfig,
  OpenDirectory,
  SaveConfig,
} from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'
import { FolderOpenOutlined } from '@ant-design/icons'

function FolderInput(props: {
  value: string
  onChange: (value: string) => void
}) {
  const [value, setValue] = useState(props.value)
  return (
    <Input
      value={value}
      addonAfter={
        <div
          style={{
            width: 32,
            height: 30,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            margin: '0 -12px'
          }}
          onClick={(e) => {
            OpenDirectory().then((v) => {
              if (v) {
                setValue(v)
                props.onChange(v)
              }
            })
          }}
        >
          <FolderOpenOutlined />
        </div>
      }
    />
  )
}

function App() {
  const [config, setConfig] = useState(null as main.Config | null)
  const [isEditing, setIsEditing] = useState(false)
    function reload() {
        LoadConfig().then((c) => {
          setConfig(c)
          setIsEditing(false)
        })
    }
  useEffect(() => {
    reload()
  }, [])
  if (!config) return <h1>Loading...</h1>

  const columns: TableColumnProps[] = [
    {
      title: 'Source',
      dataIndex: 'src',
      key: 'src',
      render(value, record, index) {
        return (
          <FolderInput
            value={value}
            onChange={(e) => {
              record.src = e
              setIsEditing(true)
            }}
          />
        )
      },
    },
    {
      title: 'Destination',
      dataIndex: 'dst',
      key: 'dst',
      render(value, record, index) {
        return (
          <FolderInput
            value={value}
            onChange={(e) => {
              record.dst = e
              setIsEditing(true)
            }}
          />
        )
      },
    },
  ]
  return (
    <div id="app">
      <div className="header">
        <h1>Sync folders</h1>
        <Space>
          <Button
            onClick={() => {
              reload()
            }}
          >
            Reload
          </Button>
          <Button
            disabled={false}
            onClick={() => {
              config.folder_pairs.push({
                src: '',
                dst: '',
              })
              setConfig(new main.Config(config))
              setIsEditing(true)
            }}
          >
            Add folder
          </Button>
          <Button
            disabled={!isEditing}
            onClick={() => {
              config.folder_pairs = config.folder_pairs.filter(
                (v) => v.src && v.dst
              )
              SaveConfig(config)
                .then(() => {
                  LoadConfig().then(setConfig)
                  notification.success({
                    message: 'Saved',
                  })
                })
                .catch((e) => {
                  notification.error({
                    message: 'Error',
                    description: e?.message,
                  })
                })
            }}
          >
            Save
          </Button>
        </Space>
      </div>
      <Table
        dataSource={config.folder_pairs}
        columns={columns}
        pagination={false}
      />
    </div>
  )
}

export default App
