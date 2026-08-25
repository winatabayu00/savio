import { Outlet } from 'react-router-dom';
import { useState } from 'react';
import { Footer } from './footer';
import { Header } from './header';
import { NavigationMenu } from './navigation-menu';

export function AppLayout() {
  const [navigationOpen, setNavigationOpen] = useState(false);

  return (
    <>
      <Header navigationOpen={navigationOpen} setNavigationOpen={setNavigationOpen} />
      <NavigationMenu navigationOpen={navigationOpen} setNavigationOpen={setNavigationOpen} />
      <main className="nxl-container">
        <div className="nxl-content">
          <div className="main-content">
            <Outlet />
          </div>
          <Footer />
        </div>
      </main>
    </>
  );
}