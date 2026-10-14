// In-memory examples for a reference prototype. No API or persistence implied.
export const fixtureCourses=[
{id:'c1',name:'Основи UI/UX дизайну',category:'Дизайн',status:'Published',students:128,price:2490,lessons:12,updated:'14.10.2026',cover:'blue',icon:'course'},
{id:'c2',name:'Таро: мова символів',category:'Символи й традиції',status:'Published',students:86,price:1890,lessons:8,updated:'13.10.2026',cover:'lime',icon:'courses'},
{id:'c3',name:'Руни: від історії до практики',category:'Символи й традиції',status:'Draft',students:0,price:2200,lessons:10,updated:'12.10.2026',cover:'neutral',icon:'lesson'},
{id:'c4',name:'Типографіка для інтерфейсів',category:'Дизайн',status:'Published',students:64,price:1490,lessons:6,updated:'11.10.2026',cover:'blue',icon:'text'},
{id:'c5',name:'Стави: композиція та значення',category:'Символи й традиції',status:'Draft',students:0,price:1750,lessons:7,updated:'10.10.2026',cover:'lime',icon:'row'},
{id:'c6',name:'Замовляння: традиція й контекст',category:'Символи й традиції',status:'Archived',students:42,price:1690,lessons:5,updated:'09.10.2026',cover:'neutral',icon:'task'}];
export const fixtureUsers=[
{id:'u1',name:'Анна Коваль',initials:'АК',email:'anna.koval@example.com',role:'Student',status:'Active',updated:'14.10.2026'},
{id:'u2',name:'Олена Коваль',initials:'ОК',email:'olena.koval@example.com',role:'Admin',status:'Active',updated:'14.10.2026'},
{id:'u3',name:'Максим Бондар',initials:'МБ',email:'max.bondar@example.com',role:'Instructor',status:'Active',updated:'13.10.2026'},
{id:'u4',name:'Софія Мельник',initials:'СМ',email:'sofia.melnyk@example.com',role:'Student',status:'Active',updated:'12.10.2026'},
{id:'u5',name:'Дмитро Савчук',initials:'ДС',email:'dmytro.savchuk@example.com',role:'Student',status:'Blocked',updated:'11.10.2026'},
{id:'u6',name:'Ірина Шевченко',initials:'ІШ',email:'iryna.shevchenko@example.com',role:'Instructor',status:'Active',updated:'10.10.2026'},
{id:'u7',name:'Артем Лисенко',initials:'АЛ',email:'artem.lysenko@example.com',role:'Student',status:'Active',updated:'09.10.2026'}];
export const fixtureLearners=[
{id:'l1',name:'Анна Коваль',initials:'АК',email:'anna.koval@example.com',course:'Основи UI/UX дизайну',courseId:'c1',status:'Active',progress:68,last:'14.10.2026 · 08:42'},
{id:'l2',name:'Софія Мельник',initials:'СМ',email:'sofia.melnyk@example.com',course:'Таро: мова символів',courseId:'c2',status:'Active',progress:42,last:'13.10.2026 · 19:15'},
{id:'l3',name:'Артем Лисенко',initials:'АЛ',email:'artem.lysenko@example.com',course:'Типографіка для інтерфейсів',courseId:'c4',status:'Active',progress:100,last:'13.10.2026 · 12:08'},
{id:'l4',name:'Дмитро Савчук',initials:'ДС',email:'dmytro.savchuk@example.com',course:'Основи UI/UX дизайну',courseId:'c1',status:'Blocked',progress:18,last:'11.10.2026 · 16:30'}];
export const fixtureReports=[
{id:'RPT-024',subject:'Некоректна поведінка в обговоренні',person:'Анна Коваль',participant:'Дмитро Савчук',course:'Основи UI/UX дизайну',description:'Учасник залишив образливий коментар під практичним завданням. Прошу перевірити обговорення та нагадати про правила спільноти.',status:'Open',date:'14.10.2026 · 08:12'},
{id:'RPT-023',subject:'Посилання на сторонню рекламу',person:'Софія Мельник',participant:'Приклад учасника',course:'Таро: мова символів',description:'У коментарях до другого уроку з’явилося рекламне посилання, яке не стосується матеріалів курсу.',status:'Open',date:'13.10.2026 · 20:05'},
{id:'RPT-022',subject:'Дубльований коментар',person:'Артем Лисенко',participant:'Приклад учасника',course:'Типографіка для інтерфейсів',description:'Повторювані коментарі видалено. Учаснику пояснено правила публікації.',status:'Resolved',date:'12.10.2026 · 10:40'}];
export const fixtureRefunds=[
{id:'RF-018',person:'Софія Мельник',email:'sofia.melnyk@example.com',course:'Таро: мова символів',amount:1890,reason:'Курс придбано помилково. Студентка ще не почала навчання та просить повернути оплату.',status:'Pending',date:'14.10.2026 · 07:50'},
{id:'RF-017',person:'Артем Лисенко',email:'artem.lysenko@example.com',course:'Основи UI/UX дизайну',amount:2490,reason:'Повторна оплата того самого курсу. Запит потребує перевірки транзакції.',status:'Pending',date:'13.10.2026 · 18:20'},
{id:'RF-016',person:'Анна Коваль',email:'anna.koval@example.com',course:'Типографіка для інтерфейсів',amount:1490,reason:'Приклад завершеного звернення для демонстрації станів черги.',status:'Refunded',date:'12.10.2026 · 11:04'}];
