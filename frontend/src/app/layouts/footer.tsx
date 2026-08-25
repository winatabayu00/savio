export function Footer() {
  return (
    <footer className="footer">
      <p className="fs-11 text-muted fw-medium text-uppercase mb-0 copyright">
        <span>Copyright ©</span>
        {new Date().getFullYear()} Savio
      </p>
      <div className="d-flex align-items-center gap-4">
        <a href="#" className="fs-11 fw-semibold text-uppercase">
          Help
        </a>
        <a href="#" className="fs-11 fw-semibold text-uppercase">
          Terms
        </a>
        <a href="#" className="fs-11 fw-semibold text-uppercase">
          Privacy
        </a>
      </div>
    </footer>
  );
}