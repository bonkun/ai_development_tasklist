import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom'; // 修正: useNavigateをインポート
import './App.css';
const Login = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate(); // 修正: useNavigateからnavigateを取得

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch('http://localhost:8080/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ "username": username, "password": password }),
      });

      const data = await response.json();

      if (data.success) {
        // トークンを保存
        localStorage.setItem('token', data.token);
        // ログイン成功時、カンバン画面に遷移
        navigate('/kanban');
      } else {
        // ログイン失敗時、エラーメッセージを表示
        setError(data.message || 'ログインに失敗しました');
      }
    } catch (error) {
      setError('サーバーエラーが発生しました');
    }
  };

  return (
    <div className="login-container">
      <div className="login-form">
        <h2>ログイン</h2>
        <form onSubmit={handleLogin}>
          <div>
            <label>ユーザー名:</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>
          <div>
            <label>パスワード:</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          {error && <p>{error}</p>}
          <button type="submit">ログイン</button>
        </form>
      </div>
    </div>
  );
};

export default Login;