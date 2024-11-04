import React from 'react';
import { BrowserRouter as Router, Route, Routes, useLocation } from 'react-router-dom';
import { CSSTransition, TransitionGroup } from 'react-transition-group';
import Login from './Login';
import Kanban from './Kanban';

const App = () => {
  return (
    <Router>
      <RouteWrapper />
    </Router>
  );
};

const RouteWrapper = () => {
  const location = useLocation();

  return (
    <TransitionGroup>
      <CSSTransition
        key={location.key}
        classNames="fade"
        timeout={300}
      >
        <Routes location={location}>
          <Route path="/login" element={<Login />} />
          <Route path="/kanban" element={<Kanban />} />
          <Route path="/" element={<Login />} />
        </Routes>
      </CSSTransition>
    </TransitionGroup>
  );
};

export default App;