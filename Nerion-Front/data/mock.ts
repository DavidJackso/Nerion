// ── Spaces ────────────────────────────────────────────────────
export interface MockSpace {
  id: number
  name: string
  slug: string
  tables: number
  members: number
  updated: string
  icon: string
  color: string
  fg: string
}

export const SPACES: MockSpace[] = [
  { id: 1, name: 'Кафедра математики',     slug: 'math-dept',      tables: 8,  members: 4, updated: '2 ч назад',    icon: 'table', color: 'var(--purple-100)', fg: 'var(--purple-700)' },
  { id: 2, name: 'Кафедра физики',         slug: 'physics-dept',   tables: 12, members: 6, updated: 'вчера',        icon: 'box',   color: '#FFE4E6',            fg: '#9F1239' },
  { id: 3, name: 'Деканат ФПМИ',           slug: 'dean-fpmi',      tables: 5,  members: 2, updated: '3 дня назад',  icon: 'users', color: '#DBEAFE',            fg: '#1E40AF' },
  { id: 4, name: 'Приёмная комиссия 2026', slug: 'admission-2026', tables: 3,  members: 8, updated: 'неделю назад', icon: 'file',  color: '#D1FAE5',            fg: 'var(--green-700)' },
]

// ── Schema ────────────────────────────────────────────────────
export interface MockTemplate {
  id: string
  icon: string
  title: string
  desc: string
  fields: number
  tag: string | null
}

export const TEMPLATES: MockTemplate[] = [
  { id: 'teachers', icon: 'users',  title: 'Преподаватели', desc: 'Учёная степень, ставка, нагрузка',  fields: 11, tag: '273-ФЗ' },
  { id: 'courses',  icon: 'file',   title: 'Курсы',         desc: 'Дисциплина, часы, форма контроля',  fields: 9,  tag: '273-ФЗ' },
  { id: 'clients',  icon: 'box',    title: 'Клиенты',       desc: 'Имя, контакты, источник',           fields: 7,  tag: null },
  { id: 'leads',    icon: 'table',  title: 'Заявки',        desc: 'Контакт, статус воронки, сумма',    fields: 8,  tag: null },
  { id: 'staff',    icon: 'users',  title: 'Сотрудники',    desc: 'ФИО, должность, контакты',          fields: 6,  tag: null },
  { id: 'blank',    icon: 'plus',   title: 'С нуля',        desc: 'Пустая таблица — определи поля сам', fields: 0, tag: null },
]

export const FIELD_TYPES: Record<string, string> = {
  text: 'Текст', number: 'Число', date: 'Дата', bool: 'Булево',
  select: 'Выбор', email: 'Email', file: '📎 Файл', relation: '↔ Связь',
}

export const FILE_ACCEPT: [string, string][] = [
  ['img', 'Изображения'], ['pdf', 'PDF'], ['doc', 'Документы'], ['video', 'Видео'], ['any', 'Любые файлы'],
]

export interface RelTable {
  name: string
  count: number
}

export const REL_TABLES: RelTable[] = [
  { name: 'Кафедры',       count: 6   },
  { name: 'Предметы',      count: 28  },
  { name: 'Курсы',         count: 42  },
  { name: 'Учебные планы', count: 8   },
  { name: 'Студенты',      count: 214 },
]

export interface InitialField {
  name: string
  type: string
  req: boolean
  accept?: string
  many?: boolean
  target?: string
}

export const INITIAL_FIELDS: InitialField[] = [
  { name: 'ФИО',            type: 'text',     req: true  },
  { name: 'Учёная степень', type: 'select',   req: true  },
  { name: 'Должность',      type: 'text',     req: true  },
  { name: 'Ставка',         type: 'number',   req: true  },
  { name: 'Email',          type: 'email',    req: false },
  { name: 'Фото',           type: 'file',     req: false, accept: 'img', many: false },
  { name: 'Дата приёма',    type: 'date',     req: true  },
  { name: 'Кафедра',        type: 'relation', req: true,  target: 'Кафедры',  many: false },
  { name: 'Предметы',       type: 'relation', req: false, target: 'Предметы', many: true  },
  { name: 'Активен',        type: 'bool',     req: true  },
]

// ── Data tables ───────────────────────────────────────────────
interface FileRef { name: string; type: string }
const FF = (name: string, type: string): FileRef => ({ name, type })

export interface TableDataEntry {
  title: string
  slug: string
  count: number
  cols: string[]
  rows: Array<Array<string | number | boolean | FileRef[]>>
}

export const TABLE_DATA: Record<string, TableDataEntry> = {
  'data-prep': {
    title: 'Преподаватели', slug: 'teachers', count: 14,
    cols: ['ФИО', 'Степень', 'Должность', 'Ставка', 'Email', 'Активен'],
    rows: [
      ['Иванов Алексей Петрович',     'д.ф.-м.н.', 'профессор', '1.00', 'ivanov@msu.ru',     true ],
      ['Петрова Мария Сергеевна',     'к.ф.-м.н.', 'доцент',    '1.00', 'petrova@msu.ru',    true ],
      ['Сидоров Дмитрий Иванович',    'к.ф.-м.н.', 'доцент',    '0.50', 'sidorov@msu.ru',    true ],
      ['Кузнецова Ольга Викторовна',  'к.т.н.',    'ст. преп.', '1.00', 'kuznetsova@msu.ru', true ],
      ['Морозов Игорь Александрович', '—',         'ассистент', '0.25', 'morozov@msu.ru',    true ],
      ['Васильев Сергей Николаевич',  'д.ф.-м.н.', 'профессор', '1.00', 'vasiliev@msu.ru',   true ],
      ['Новикова Анна Дмитриевна',    'к.ф.-м.н.', 'доцент',    '1.00', 'novikova@msu.ru',   true ],
      ['Орлов Павел Владимирович',    '—',         'ассистент', '0.50', 'orlov@msu.ru',      true ],
      ['Соколова Елена Михайловна',   'к.п.н.',    'доцент',    '1.00', 'sokolova@msu.ru',   false],
      ['Лебедев Артём Романович',     'к.ф.-м.н.', 'доцент',    '1.00', 'lebedev@msu.ru',    true ],
      ['Григорьев Михаил Олегович',   'д.ф.-м.н.', 'профессор', '1.00', 'grigoriev@msu.ru',  true ],
      ['Тихонова Светлана Ильинична', 'к.ф.-м.н.', 'доцент',    '0.75', 'tikhonova@msu.ru',  true ],
      ['Андреев Виктор Юрьевич',      '—',         'ст. преп.', '1.00', 'andreev@msu.ru',    true ],
      ['Смирнов Денис Геннадьевич',   'к.ф.-м.н.', 'доцент',    '1.00', 'smirnov@msu.ru',    true ],
    ],
  },
  'data-courses': {
    title: 'Курсы', slug: 'courses', count: 42,
    cols: ['Название', 'Кафедра', 'Форма контроля', 'Часов', 'Материалы', 'Активен'],
    rows: [
      ['Математический анализ',  'Кафедра математики', 'Экзамен', 144, [FF('Лекция_1.jpg','img'),FF('График.png','img'),FF('Силлабус.pdf','pdf'),FF('Лекции.pdf','pdf'),FF('Ряды.pptx','ppt')], true ],
      ['Линейная алгебра',       'Кафедра математики', 'Экзамен', 72,  [FF('Матрицы.png','img'),FF('Силлабус.pdf','pdf'),FF('Задачник.pdf','pdf')], true ],
      ['Дискретная математика',  'Кафедра математики', 'Зачёт',   54,  [FF('Граф.png','img'),FF('Дерево.png','img'),FF('Конспект.pdf','pdf'),FF('Вводная.mp4','video')], true ],
      ['Теория вероятностей',    'Кафедра математики', 'Экзамен', 72,  [FF('Силлабус.pdf','pdf'),FF('Байес.pptx','ppt')], true ],
      ['Численные методы',       'Кафедра математики', 'Зачёт',   54,  [FF('Сходимость.png','img'),FF('Силлабус.pdf','pdf'),FF('Код.zip','zip')], true ],
      ['Алгебра',                'Кафедра математики', 'Экзамен', 108, [FF('Группы.png','img'),FF('Силлабус.pdf','pdf')], true ],
      ['Геометрия',              'Кафедра математики', 'Экзамен', 72,  [FF('Чертёж_1.png','img'),FF('Чертёж_2.png','img'),FF('Поверхность.png','img'),FF('Силлабус.pdf','pdf')], true ],
      ['Уравнения мат. физики',  'Кафедра математики', 'Экзамен', 72,  [FF('Фронт.png','img'),FF('Лекции.pdf','pdf')], true ],
      ['Топология',              'Кафедра математики', 'Зачёт',   36,  [FF('Мёбиус.png','img')], false],
      ['Программирование',       'Кафедра АСОИУ',      'Зачёт',   72,  [FF('Исходники.zip','zip'),FF('Методичка.pdf','pdf')], true ],
      ['Физика',                 'Кафедра физики',     'Экзамен', 144, [FF('Установка.jpg','img'),FF('Опыт.jpg','img'),FF('Лабы.pdf','pdf'),FF('Демо.mp4','video')], true ],
      ['Статистика',             'Кафедра математики', 'Зачёт',   54,  [FF('Гистограмма.png','img'),FF('Силлабус.pdf','pdf')], true ],
    ],
  },
  'data-plans': {
    title: 'Учебные планы', slug: 'plans', count: 8,
    cols: ['Специальность', 'Направление', 'Форма', 'Срок', 'Год набора', 'Статус'],
    rows: [
      ['01.03.01 Математика',            'Прикладная математика',            'очная',      '4 года', '2022', 'active' ],
      ['01.03.01 Математика',            'Математика и компьютерные науки',  'очная',      '4 года', '2023', 'active' ],
      ['01.03.02 Прикладная математика', 'Вычислительная математика',        'очная',      '4 года', '2022', 'active' ],
      ['09.03.01 Информатика и ВТ',      'Программная инженерия',            'очная',      '4 года', '2024', 'active' ],
      ['01.04.01 Математика',            'Современная математика',           'очная',      '2 года', '2024', 'active' ],
      ['01.03.01 Математика',            'Прикладная математика',            'заочная',    '4 года', '2022', 'setup'  ],
      ['09.04.01 Информатика',           'Математические методы в IT',       'очная',      '2 года', '2025', 'setup'  ],
      ['01.03.02 Прикладная математика', 'Математическое моделирование',     'очно-заочн', '4 года', '2023', 'offline'],
    ],
  },
}

export interface ColMeta {
  kind: string
  options?: string[]
  relTable?: string
  many?: boolean
  mono?: boolean
  placeholder?: string
  single?: boolean
  accept?: string
}

export const COL_META: Record<string, ColMeta> = {
  'Степень':        { kind: 'select',   options: ['д.ф.-м.н.', 'к.ф.-м.н.', 'к.т.н.', 'к.п.н.', '—'] },
  'Должность':      { kind: 'select',   options: ['профессор', 'доцент', 'ст. преп.', 'ассистент'] },
  'Кафедра':        { kind: 'relation', relTable: 'Кафедры',  many: false, options: ['Математического анализа', 'Высшей алгебры', 'Дифференциальной геометрии', 'Вычислительной математики'] },
  'Форма контроля': { kind: 'select',   options: ['Экзамен', 'Зачёт', 'Дифф. зачёт'] },
  'Форма':          { kind: 'select',   options: ['очная', 'заочная', 'очно-заочная'] },
  'Активен':        { kind: 'bool' },
  'Email':          { kind: 'text',     mono: true, placeholder: 'name@msu.ru' },
  'Ставка':         { kind: 'select',   options: ['1.00', '0.75', '0.50', '0.25'] },
  'Часов':          { kind: 'number' },
  'Материалы':      { kind: 'file' },
  'Фото':           { kind: 'file', single: true, accept: 'img' },
}

// ── API ───────────────────────────────────────────────────────
export interface MockEndpoint {
  method: string
  path: string
  desc: string
}

export const ENDPOINTS: MockEndpoint[] = [
  { method: 'GET',    path: '/api/math-dept/teachers',      desc: 'Список с фильтрами и пагинацией' },
  { method: 'GET',    path: '/api/math-dept/teachers/{id}', desc: 'Одна запись по ID' },
  { method: 'POST',   path: '/api/math-dept/teachers',      desc: 'Создать запись' },
  { method: 'PATCH',  path: '/api/math-dept/teachers/{id}', desc: 'Обновить поля' },
  { method: 'DELETE', path: '/api/math-dept/teachers/{id}', desc: 'Удалить запись' },
]

export interface MockApiKey {
  name: string
  key: string
  scope: string
  last: string
  req: number
}

export const API_KEYS: MockApiKey[] = [
  { name: 'Production · LMS',     key: 'nrn_live_8a72f9b1c4e5d6a3f2a', scope: 'read+write', last: '2 мин назад', req: 14820 },
  { name: 'Аналитика (R/O)',      key: 'nrn_live_3c91e2a7f4b8d6c1e2b', scope: 'read',       last: '1 ч назад',  req: 312   },
  { name: 'Локальная разработка', key: 'nrn_test_•••••••••••••••••••',  scope: 'read+write', last: 'вчера',      req: 47    },
]

// ── Team & Settings ────────────────────────────────────────────
export interface MockMember {
  name: string
  email: string
  role: string
  joined: string
  color: string
  fg: string
  you?: boolean
}

export const MEMBERS: MockMember[] = [
  { name: 'Анна Иванова',   email: 'anna.ivanova@msu.ru', role: 'Admin',  joined: '12.03.2024', color: 'var(--purple-200)', fg: 'var(--purple-700)', you: true },
  { name: 'Дмитрий Петров', email: 'd.petrov@msu.ru',    role: 'Member', joined: '15.03.2024', color: '#DBEAFE',            fg: '#1E40AF' },
  { name: 'Мария Сидорова', email: 'm.sidorova@msu.ru',  role: 'Member', joined: '20.03.2024', color: '#FFE4E6',            fg: '#9F1239' },
  { name: 'Сергей Ковалёв', email: 's.kovalev@msu.ru',   role: 'Member', joined: '01.04.2024', color: '#D1FAE5',            fg: 'var(--green-700)' },
]

// ── PDF ───────────────────────────────────────────────────────
export interface MockPdfTemplate {
  id: string
  name: string
  gost: string
  table: string
  fields: number
  bound: number
  docs: number
  last: string
  status: string
}

export const MY_TEMPLATES: MockPdfTemplate[] = [
  { id: 'spravka',  name: 'Справка с места работы',    gost: 'ГОСТ Р 7.0.97-2016',      table: 'Преподаватели',   fields: 7,  bound: 7,  docs: 184, last: '2 ч назад',    status: 'ready' },
  { id: 'vedomost', name: 'Ведомость аттестации',      gost: 'Приказ Минобрнауки № 301', table: 'Студенты · Курсы',fields: 9,  bound: 9,  docs: 312, last: 'вчера',        status: 'ready' },
  { id: 'nagruzka', name: 'Учебная нагрузка',          gost: 'ГОСТ Р 7.0.97-2016',      table: 'Преподаватели',   fields: 12, bound: 10, docs: 47,  last: '3 дня назад',  status: 'setup' },
  { id: 'prikaz',   name: 'Приказ о приёме на работу', gost: 'ТК РФ ст. 68',            table: 'Преподаватели',   fields: 11, bound: 11, docs: 42,  last: 'неделю назад', status: 'ready' },
]

export interface LibCat {
  id: string
  label: string
  count: number
}

export const LIB_CATS: LibCat[] = [
  { id: 'all',  label: 'Все',             count: 47 },
  { id: 'edu',  label: 'Учебная работа',  count: 18 },
  { id: 'hr',   label: 'Кадры',           count: 11 },
  { id: 'stud', label: 'Студенты',        count: 9  },
]

export interface LibTemplate {
  cat: string
  name: string
  gost: string
  desc: string
  used: number
  official: boolean
}

export const LIB_TPLS: LibTemplate[] = [
  { cat: 'edu',  name: 'Расписание занятий группы',   gost: 'Локальный шаблон',          desc: 'Недельная сетка по дисциплинам и аудиториям.',  used: 89,  official: false },
  { cat: 'edu',  name: 'Индивидуальный учебный план', gost: 'ФЗ-273 ст. 34',             desc: 'ИУП студента с дисциплинами и сроками.',        used: 31,  official: true  },
  { cat: 'hr',   name: 'Должностная инструкция',      gost: 'ГОСТ Р 7.0.97-2016',        desc: 'Преамбула, обязанности, права.',                used: 28,  official: false },
  { cat: 'hr',   name: 'Карточка сотрудника Т-2',     gost: 'Постановление Госкомстата', desc: 'Личная карточка по форме.',                     used: 14,  official: true  },
  { cat: 'stud', name: 'Справка об обучении',         gost: 'Приказ Минобрнауки № 455',  desc: 'ФИО, группа, форма, направление.',              used: 168, official: true  },
  { cat: 'stud', name: 'Академическая справка',       gost: 'Приказ Минобрнауки № 455',  desc: 'Курсы с оценками за весь период.',              used: 47,  official: true  },
  { cat: 'fin',  name: 'Отчёт по часам кафедры',      gost: 'Локальный шаблон',          desc: 'Суммарная нагрузка кафедры.',                   used: 12,  official: false },
]

export const PDF_RECORDS: string[] = [
  'Иванов Алексей Петрович', 'Петрова Мария Сергеевна', 'Сидоров Дмитрий Иванович',
  'Кузнецова Ольга Викторовна', 'Васильев Сергей Николаевич', 'Орлов Павел Владимирович',
]

export const PDF_FORMATS: string[] = [
  '—', 'именительный', 'родительный', 'дательный', 'длинная дата', 'короткая дата', 'инициалы', 'прописью',
]

export const BIND_TABLES: Record<string, string[]> = {
  'Преподаватели': ['ФИО', 'Степень', 'Должность', 'Ставка', 'Email', 'Дата приёма', 'Кафедра'],
  'Организация':   ['Министерство', 'Полное название', 'Краткое название'],
  'Подписант':     ['Должность', 'ФИО'],
  'Система':       ['Дата выдачи', 'Номер документа'],
}

export interface MappingEntry {
  ph: string
  label: string
  table: string
  field: string
  fmt: string
  sample: string
}

export const INITIAL_MAPPING: MappingEntry[] = [
  { ph: 'fio',        label: 'ФИО получателя',  table: 'Преподаватели', field: 'ФИО',              fmt: 'дательный',    sample: 'Иванову А. П.' },
  { ph: 'position',   label: 'Должность',       table: 'Преподаватели', field: 'Должность',        fmt: '—',            sample: 'профессор' },
  { ph: 'kafedra',    label: 'Кафедра',         table: 'Преподаватели', field: 'Кафедра',          fmt: 'родительный',  sample: 'математического анализа' },
  { ph: 'hired_at',   label: 'Дата приёма',     table: 'Преподаватели', field: 'Дата приёма',      fmt: 'длинная дата', sample: '1 сентября 2008 г.' },
  { ph: 'rate',       label: 'Ставка',          table: 'Преподаватели', field: 'Ставка',           fmt: 'прописью',     sample: 'полная' },
  { ph: 'org',        label: 'Организация',     table: 'Организация',   field: 'Краткое название', fmt: '—',            sample: 'МГУ' },
  { ph: 'today',      label: 'Дата выдачи',     table: 'Система',       field: 'Дата выдачи',      fmt: 'длинная дата', sample: '15 сентября 2025 г.' },
  { ph: 'signer_fio', label: 'Подписант',       table: '',              field: '',                 fmt: 'инициалы',     sample: '—' },
]

export interface PdfArchiveEntry {
  kind: string
  name: string
  tpl: string
  rec: string
  date: string
  size: string
  by: string
}

export const PDF_ARCHIVE: PdfArchiveEntry[] = [
  { kind: 'single', name: 'Справка_Иванов_АП.pdf',         tpl: 'Справка с места работы', rec: 'Иванов Алексей Петрович',  date: 'Сегодня, 14:32',  size: '142 КБ', by: 'Анна Иванова' },
  { kind: 'bulk',   name: 'Ведомости_аттестации_318М.zip', tpl: 'Ведомость аттестации',   rec: '28 студентов',             date: 'Сегодня, 11:05',  size: '3.4 МБ', by: 'Дмитрий Петров' },
  { kind: 'single', name: 'Справка_Петрова_МС.pdf',        tpl: 'Справка с места работы', rec: 'Петрова Мария Сергеевна',  date: 'Вчера, 16:48',    size: '139 КБ', by: 'Анна Иванова' },
  { kind: 'bulk',   name: 'Учебная_нагрузка_кафедра.zip',  tpl: 'Учебная нагрузка',       rec: '14 преподавателей',        date: '18 июня, 09:12',  size: '1.8 МБ', by: 'Анна Иванова' },
  { kind: 'single', name: 'Приказ_184к_Орлов.pdf',         tpl: 'Приказ о приёме',        rec: 'Орлов Павел Владимирович', date: '15 июня, 13:20',  size: '98 КБ',  by: 'Анна Иванова' },
]

// ── Files ─────────────────────────────────────────────────────
export interface MockFile {
  name: string
  type: string
  kb: number
  src: string
  tpl?: string
  course?: string
  date: string
}

export const F_FILES: MockFile[] = [
  { name: 'Справка_Иванов_АП.pdf',    type: 'pdf',   kb: 142,    src: 'gen',    tpl: 'Справка',   date: 'Сегодня, 14:32' },
  { name: 'Справка_Петрова_МС.pdf',   type: 'pdf',   kb: 139,    src: 'gen',    tpl: 'Справка',   date: 'Сегодня, 14:31' },
  { name: 'Справка_Сидоров_ДИ.pdf',   type: 'pdf',   kb: 141,    src: 'gen',    tpl: 'Справка',   date: 'Сегодня, 14:31' },
  { name: 'Справка_Кузнецова_ОВ.pdf', type: 'pdf',   kb: 140,    src: 'gen',    tpl: 'Справка',   date: 'Сегодня, 14:30' },
  { name: 'Справка_Орлов_ПВ.pdf',     type: 'pdf',   kb: 138,    src: 'gen',    tpl: 'Справка',   date: 'Сегодня, 14:30' },
  { name: 'Ведомость_318М.pdf',       type: 'pdf',   kb: 264,    src: 'gen',    tpl: 'Ведомость', date: 'Вчера, 11:05' },
  { name: 'Ведомость_412М.pdf',       type: 'pdf',   kb: 258,    src: 'gen',    tpl: 'Ведомость', date: 'Вчера, 11:04' },
  { name: 'Приказ_184к_Орлов.pdf',    type: 'pdf',   kb: 98,     src: 'gen',    tpl: 'Приказ',    date: '18 июня' },
  { name: 'Приказ_185к_Андреев.pdf',  type: 'pdf',   kb: 96,     src: 'gen',    tpl: 'Приказ',    date: '18 июня' },
  { name: 'Нагрузка_кафедра.pdf',     type: 'pdf',   kb: 412,    src: 'gen',    tpl: 'Нагрузка',  date: '15 июня' },
  { name: 'Лекция_1_доска.jpg',       type: 'img',   kb: 1840,   src: 'course', course: 'Математический анализ', date: '16 июня' },
  { name: 'График_ряда.png',          type: 'img',   kb: 320,    src: 'course', course: 'Математический анализ', date: '16 июня' },
  { name: 'Силлабус_МатАнализ.pdf',   type: 'pdf',   kb: 248,    src: 'course', course: 'Математический анализ', date: '16 июня' },
  { name: 'Матрицы_схема.png',        type: 'img',   kb: 412,    src: 'course', course: 'Линейная алгебра',      date: '15 июня' },
  { name: 'Граф_пример.png',          type: 'img',   kb: 286,    src: 'course', course: 'Дискретная математика', date: '12 июня' },
  { name: 'Дерево_обхода.png',        type: 'img',   kb: 244,    src: 'course', course: 'Дискретная математика', date: '12 июня' },
  { name: 'Чертёж_1.png',            type: 'img',   kb: 520,    src: 'course', course: 'Геометрия',             date: '10 июня' },
  { name: 'Чертёж_2.png',            type: 'img',   kb: 540,    src: 'course', course: 'Геометрия',             date: '10 июня' },
  { name: 'Код_примеры.zip',          type: 'zip',   kb: 1240,   src: 'course', course: 'Численные методы',      date: '5 июня' },
  { name: 'Лабы.pdf',                 type: 'pdf',   kb: 1620,   src: 'course', course: 'Физика',                date: '3 июня' },
  { name: 'Демонстрация.mp4',         type: 'video', kb: 184320, src: 'course', course: 'Физика',                date: '3 июня' },
  { name: 'Логотип_кафедры.png',      type: 'img',   kb: 88,     src: 'upload', date: '1 июня' },
]

export interface MockList {
  slug: string
  name: string
  icon: string
  rule: [string, string]
  pred: (f: MockFile) => boolean
  public: boolean
  hero?: boolean
}

export const F_LISTS_INITIAL: MockList[] = [
  { slug: 'generated-docs', name: 'Документы из генератора', icon: 'file',   rule: ['Источник', 'генерация'], pred: (f) => f.src === 'gen',     public: true,  hero: true },
  { slug: 'spravki',        name: 'Справки с места работы',  icon: 'file',   rule: ['Шаблон', 'Справка'],     pred: (f) => f.tpl === 'Справка', public: true },
  { slug: 'materials',      name: 'Учебные материалы',       icon: 'folder', rule: ['Источник', 'курсы'],     pred: (f) => f.src === 'course',  public: false },
  { slug: 'gallery',        name: 'Галерея изображений',     icon: 'image',  rule: ['Тип', 'изображение'],    pred: (f) => f.type === 'img',    public: true },
]

export interface FileTone {
  label: string
  bg: string
  fg: string
  bd: string
}

export const FILE_TONE: Record<string, FileTone> = {
  img:   { label: 'IMG', bg: 'var(--brand-tint)',  fg: 'var(--purple-600)',  bd: 'var(--purple-100)'      },
  pdf:   { label: 'PDF', bg: 'var(--red-50)',      fg: 'var(--red-600)',     bd: 'var(--red-100)'         },
  doc:   { label: 'DOC', bg: 'var(--blue-50)',     fg: 'var(--blue-600)',    bd: '#DBEAFE'                },
  ppt:   { label: 'PPT', bg: 'var(--amber-50)',    fg: 'var(--amber-700)',   bd: 'var(--amber-100)'       },
  pptx:  { label: 'PPT', bg: 'var(--amber-50)',    fg: 'var(--amber-700)',   bd: 'var(--amber-100)'       },
  video: { label: 'MP4', bg: 'var(--purple-50)',   fg: 'var(--purple-700)', bd: 'var(--purple-100)'      },
  zip:   { label: 'ZIP', bg: 'var(--neutral-100)', fg: 'var(--neutral-600)',bd: 'var(--border-default)'  },
  docx:  { label: 'DOC', bg: 'var(--blue-50)',     fg: 'var(--blue-600)',    bd: '#DBEAFE'                },
}
