import { useEffect, useState } from 'react'
import './App.css'
import {
  Button,
  Input,
  notification,
  Popconfirm,
  Space,
  Table,
  type TableColumnProps,
} from 'antd'
import {
  LoadConfig,
  OpenDirectory,
  ReadLogs,
  SaveConfig,
} from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'
import { WindowSetDarkTheme } from '../wailsjs/runtime/runtime'
import { FolderOpenOutlined } from '@ant-design/icons'
import { Immer } from 'immer'

const immer = new Immer()

function FolderInput(props: {
  value: string
  onChange: (value: string) => void
}) {
  const [value, setValue] = useState(props.value)
  return (
    <Input
      value={value}
      onChange={(e) => {
        setValue(e.target.value)
        props.onChange(e.target.value)
      }}
      addonAfter={
        <div
          style={{
            width: 32,
            height: 30,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            margin: '0 -12px',
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
  const [logs, setLogs] = useState([] as string[])
  function reload() {
    LoadConfig().then((c) => {
      setConfig(c)
      setIsEditing(false)
    })
  }
  useEffect(() => {
    reload()
    ReadLogs().then(setLogs)
    setInterval(() => {
      ReadLogs().then(setLogs)
    }, 5000)
    WindowSetDarkTheme()
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
              setIsEditing(true)
              config.folder_pairs = immer.produce(config.folder_pairs, (fp) => {
                fp[index].src = e
              })
              setConfig(new main.Config(config))
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
              setIsEditing(true)
              config.folder_pairs = immer.produce(config.folder_pairs, (fp) => {
                fp[index].dst = e
              })
              setConfig(new main.Config(config))
            }}
          />
        )
      },
    },
    {
      title: 'Action',
      dataIndex: 'Action',
      key: 'Action',
      width: 20,
      render(value, record, index) {
        return (
          <Popconfirm
            title="Are you sure to remove?"
            onConfirm={(e) => {
              config.folder_pairs = config.folder_pairs.filter(
                (v, i) => i !== index
              )
              setConfig(new main.Config(config))
              setIsEditing(true)
            }}
          >
            <Button size="small" color="red" variant="outlined">
              Del
            </Button>
          </Popconfirm>
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
            Add
          </Button>
          <Button
            disabled={!isEditing}
            onClick={() => {
              config.folder_pairs = config.folder_pairs.filter(
                (v) => v.src && v.dst
              )
              SaveConfig(config)
                .then(() => {
                  reload()
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
        key={config.folder_pairs.length}
        dataSource={config.folder_pairs}
        columns={columns}
        pagination={false}
      />
      <div className="">
        <h2>Logs</h2>
        <div className="logs">
          {logs.slice().reverse().map((v, i) => (
            <div key={i}>{v}</div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default App
