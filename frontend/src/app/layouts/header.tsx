import { useEffect, useRef, useState } from 'react';
import {
  FiAlignLeft,
  FiArrowRight,
  FiLogOut,
  FiSettings,
} from 'react-icons/fi';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/app/providers/auth-provider';

interface HeaderProps {
  navigationOpen: boolean;
  setNavigationOpen: (open: boolean) => void;
}

export function Header({ navigationOpen, setNavigationOpen }: HeaderProps) {
  const { auth, logout } = useAuth();
  const navigate = useNavigate();
  const [profileOpen, setProfileOpen] = useState(false);
  const [miniMenu, setMiniMenu] = useState(false);
  const profileRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const isMini = window.matchMedia('(min-width:1025px) and (max-width:1400px)').matches;
    setMiniMenu(isMini);
    if (isMini) document.documentElement.classList.add('minimenu');
    return () => {
      document.documentElement.classList.remove('minimenu');
    };
  }, []);

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (profileRef.current && !profileRef.current.contains(e.target as Node)) {
        setProfileOpen(false);
      }
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, []);

  const toggleMini = () => {
    setMiniMenu((m) => !m);
    document.documentElement.classList.toggle('minimenu');
  };

  const toggleProfile = () => setProfileOpen((o) => !o);

  const onLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <header className="nxl-header">
      <div className="header-wrapper">
        <div className="header-left d-flex align-items-center gap-4">
          <a
            href="#"
            className="nxl-head-mobile-toggler"
            onClick={(e) => {
              e.preventDefault();
              setNavigationOpen(!navigationOpen);
            }}
            id="mobile-collapse"
          >
            <div className={`hamburger hamburger--arrowturn ${navigationOpen ? 'is-active' : ''}`}>
              <div className="hamburger-box">
                <div className="hamburger-inner"></div>
              </div>
            </div>
          </a>
          <div className="nxl-navigation-toggle">
            <a
              href="#"
              onClick={(e) => {
                e.preventDefault();
                toggleMini();
              }}
            >
              {miniMenu ? <FiArrowRight size={24} /> : <FiAlignLeft size={24} />}
            </a>
          </div>
        </div>
        <div className="header-right ms-auto">
          <div className="d-flex align-items-center">
            <div ref={profileRef} className="dropdown nxl-h-item">
              <a
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  toggleProfile();
                }}
                data-bs-auto-close="outside"
                aria-expanded={profileOpen}
                className="nxl-head-link"
              >
                <span className="user-avtar user-avtar-text">
                  {(auth?.user.name ?? 'U')
                    .split(' ')
                    .map((p) => p.charAt(0))
                    .slice(0, 2)
                    .join('')
                    .toUpperCase()}
                </span>
              </a>
              <div className={`dropdown-menu dropdown-menu-end nxl-h-dropdown nxl-user-dropdown ${profileOpen ? 'show' : ''}`}>
                <div className="dropdown-header">
                  <div className="d-flex align-items-center">
                    <div>
                      <h6 className="text-dark mb-0">
                        {auth?.user.name}
                        <span className="badge bg-soft-success text-success ms-1 text-capitalize">
                          {auth?.role ?? 'MEMBER'}
                        </span>
                      </h6>
                      <span className="fs-12 fw-medium text-muted">{auth?.user.email}</span>
                    </div>
                  </div>
                </div>
                <a href="#" className="dropdown-item" onClick={(e) => e.preventDefault()}>
                  <i>
                    <FiSettings />
                  </i>
                  <span>Account Settings</span>
                </a>
                <div className="dropdown-divider"></div>
                <a
                  href="#"
                  className="dropdown-item"
                  onClick={(e) => {
                    e.preventDefault();
                    void onLogout();
                  }}
                >
                  <i>
                    <FiLogOut />
                  </i>
                  <span>Logout</span>
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}