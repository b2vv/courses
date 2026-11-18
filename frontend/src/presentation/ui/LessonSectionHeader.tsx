export function LessonSectionHeader({ title, status }: { title: string; status: string }) {
  return (
    <div className="lesson-section-heading">
      <h4>{title}</h4>
      <p>{status}</p>
    </div>
  );
}
