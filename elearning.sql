--
-- PostgreSQL database dump
--

\restrict En0mRVaFoxH5CtcVf8wa2ADfiWdCX0jx4dIIMaq8bekEyZTXpNitEJJ3qSC43l6

-- Dumped from database version 18.1
-- Dumped by pg_dump version 18.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: enrollment_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.enrollment_status AS ENUM (
    'active',
    'completed',
    'dropped'
);


ALTER TYPE public.enrollment_status OWNER TO postgres;

--
-- Name: notification_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.notification_type AS ENUM (
    'enrollment',
    'new_lesson',
    'completed'
);


ALTER TYPE public.notification_type OWNER TO postgres;

--
-- Name: user_role; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.user_role AS ENUM (
    'student',
    'teacher',
    'admin'
);


ALTER TYPE public.user_role OWNER TO postgres;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.categories (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.categories OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.categories_id_seq OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: course_details; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.course_details AS
SELECT
    NULL::integer AS id,
    NULL::character varying(200) AS title,
    NULL::text AS description,
    NULL::character varying(255) AS thumbnail,
    NULL::boolean AS is_published,
    NULL::character varying(100) AS category_name,
    NULL::character varying(100) AS teacher_name,
    NULL::character varying(100) AS teacher_email,
    NULL::bigint AS total_lessons,
    NULL::bigint AS total_students,
    NULL::timestamp without time zone AS created_at,
    NULL::timestamp without time zone AS updated_at;


ALTER VIEW public.course_details OWNER TO postgres;

--
-- Name: courses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.courses (
    id integer NOT NULL,
    title character varying(200) NOT NULL,
    description text NOT NULL,
    thumbnail character varying(255),
    category_id integer,
    teacher_id integer NOT NULL,
    is_published boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.courses OWNER TO postgres;

--
-- Name: courses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.courses_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.courses_id_seq OWNER TO postgres;

--
-- Name: courses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.courses_id_seq OWNED BY public.courses.id;


--
-- Name: enrollments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.enrollments (
    id integer NOT NULL,
    user_id integer NOT NULL,
    course_id integer NOT NULL,
    enrolled_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status public.enrollment_status DEFAULT 'active'::public.enrollment_status NOT NULL,
    completed_at timestamp without time zone
);


ALTER TABLE public.enrollments OWNER TO postgres;

--
-- Name: enrollments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.enrollments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.enrollments_id_seq OWNER TO postgres;

--
-- Name: enrollments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.enrollments_id_seq OWNED BY public.enrollments.id;


--
-- Name: lessons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.lessons (
    id integer NOT NULL,
    course_id integer NOT NULL,
    title character varying(200) NOT NULL,
    content text,
    video_url character varying(255),
    file_url character varying(255),
    order_number integer DEFAULT 0 NOT NULL,
    duration integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.lessons OWNER TO postgres;

--
-- Name: lessons_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.lessons_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.lessons_id_seq OWNER TO postgres;

--
-- Name: lessons_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.lessons_id_seq OWNED BY public.lessons.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notifications (
    id integer NOT NULL,
    user_id integer NOT NULL,
    title character varying(200) NOT NULL,
    message text NOT NULL,
    type public.notification_type NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.notifications OWNER TO postgres;

--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notifications_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notifications_id_seq OWNER TO postgres;

--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: progress; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.progress (
    id integer NOT NULL,
    user_id integer NOT NULL,
    lesson_id integer NOT NULL,
    is_completed boolean DEFAULT false NOT NULL,
    completed_at timestamp without time zone
);


ALTER TABLE public.progress OWNER TO postgres;

--
-- Name: progress_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.progress_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.progress_id_seq OWNER TO postgres;

--
-- Name: progress_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.progress_id_seq OWNED BY public.progress.id;


--
-- Name: student_progress; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.student_progress AS
 SELECT e.user_id,
    e.course_id,
    c.title AS course_title,
    count(DISTINCT l.id) AS total_lessons,
    count(DISTINCT p.id) AS completed_lessons,
    round((((count(DISTINCT p.id))::numeric / (NULLIF(count(DISTINCT l.id), 0))::numeric) * (100)::numeric), 2) AS progress_percentage,
    e.enrolled_at,
    e.status
   FROM (((public.enrollments e
     JOIN public.courses c ON ((e.course_id = c.id)))
     LEFT JOIN public.lessons l ON ((c.id = l.course_id)))
     LEFT JOIN public.progress p ON (((l.id = p.lesson_id) AND (p.user_id = e.user_id) AND (p.is_completed = true))))
  GROUP BY e.user_id, e.course_id, c.title, e.enrolled_at, e.status;


ALTER VIEW public.student_progress OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    email character varying(100) NOT NULL,
    password character varying(255) NOT NULL,
    role public.user_role DEFAULT 'student'::public.user_role NOT NULL,
    avatar character varying(255),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: courses id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses ALTER COLUMN id SET DEFAULT nextval('public.courses_id_seq'::regclass);


--
-- Name: enrollments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.enrollments ALTER COLUMN id SET DEFAULT nextval('public.enrollments_id_seq'::regclass);


--
-- Name: lessons id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lessons ALTER COLUMN id SET DEFAULT nextval('public.lessons_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: progress id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.progress ALTER COLUMN id SET DEFAULT nextval('public.progress_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.categories (id, name, description, created_at) FROM stdin;
1	Programming	Learn programming languages and frameworks	2025-11-24 16:32:46.614112
2	Design	UI/UX, Graphic Design, and more	2025-11-24 16:32:46.614112
3	Business	Marketing, Management, Entrepreneurship	2025-11-24 16:32:46.614112
4	Data Science	Machine Learning, AI, Data Analysis	2025-11-24 16:32:46.614112
\.


--
-- Data for Name: courses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.courses (id, title, description, thumbnail, category_id, teacher_id, is_published, created_at, updated_at) FROM stdin;
4	Linux 101	a brief introduction to linux for baby.	https://upload.wikimedia.org/wikipedia/commons/thumb/3/35/Tux.svg/330px-Tux.svg.png	\N	8	t	2025-11-27 06:24:56.786266	2025-11-27 13:44:47.593365
2	Untitled	wel	https://upload.wikimedia.org/wikipedia/commons/thumb/2/23/Golang.png/960px-Golang.png	\N	6	t	2025-11-25 03:19:12.710573	2025-11-30 14:20:11.169001
\.


--
-- Data for Name: enrollments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.enrollments (id, user_id, course_id, enrolled_at, status, completed_at) FROM stdin;
1	7	2	2025-11-25 13:27:33.571966	completed	\N
3	7	4	2025-11-27 13:45:07.160844	completed	\N
4	13	4	2025-11-30 15:58:57.826558	active	\N
\.


--
-- Data for Name: lessons; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.lessons (id, course_id, title, content, video_url, file_url, order_number, duration, created_at, updated_at) FROM stdin;
9	4	Introduction	**Lesson 1 — What Even Is Linux?**\r\n\r\nAlright, welcome to the very first step of getting into Linux. If you’ve ever heard people talk about Linux like it’s some hacker-only secret club — don’t worry, it’s not that deep. You’re about to understand it without touching a single terminal yet.\r\n\r\n_So… what is Linux?_\r\n\r\nLinux is basically an operating system — like Windows or macOS — but open-source and way more customizable. Instead of being owned by one company, Linux is built and improved by a global community. Anyone can peek at the code, change it, fix it, or improve it.\r\n\r\nIt’s everywhere too:\r\n\r\nmost cloud servers run Linux\r\n\r\nAndroid phones use a Linux kernel\r\n\r\nsupercomputers, smart TVs, routers… all running Linux behind the scenes\r\n\r\nPretty wild, right? You’re learning something billions of devices depend on.\r\n\r\nWhy do people use Linux?\r\n\r\nDifferent folks, different reasons. Here are the big ones:\r\n\r\nIt’s free\r\n\r\nIt’s stable and great for long-running servers\r\n\r\nIt gives you full control over your system instead of forcing things on you\r\n\r\nFamous for being secure and private\r\n\r\nPerfect for developers, DevOps, cybersecurity, and server admins\r\n\r\nBasically, once you understand Linux, you get superpowers in tech.\r\n\r\nWhat makes Linux different from Windows?\r\n\r\nHere’s a quick vibe check:\r\n\r\nWindows\tLinux\r\nOwned by Microsoft\tOpen source, many communities\r\nLimited customization\tFully customizable\r\nMostly GUI-based\tTerminal-focused (but GUIs exist)\r\nExpensive for servers\tFree for everything\r\nCommon for personal computers\tCommon for servers/dev environments\r\n\r\nNot one is “better,” it just depends on what you’re trying to do. Linux is the go-to for tech-heavy work.\r\n\r\nWhat are "distributions" (distros)?\r\n\r\nLinux isn’t one single download. There are distros — different flavors of Linux. They’re all based on the same Linux core, but they package software and tools differently.\r\n\r\nSome popular ones:\r\n\r\nUbuntu — easiest for beginners\r\n\r\nLinux Mint — super friendly if you’re coming from Windows\r\n\r\nFedora — up-to-date but stable\r\n\r\nArch Linux — “build everything yourself” energy\r\n\r\nKali Linux — for cybersecurity folks\r\n\r\nDon’t stress about which one yet — we’ll get you picking the right distro later.\r\n\r\nWhat can you do on Linux?\r\n\r\nLiterally everything:\r\n\r\n* web development\r\n* \r\n* cybersecurity testing\r\n* \r\n* running cloud servers\r\n* \r\n* hosting websites\r\n* \r\n* Docker / Kubernetes\r\n* \r\n* gaming (yup, Steam works)\r\n* \r\n* everyday office stuff\r\n*\r\n\r\nLinux isn’t just for “nerds”. It’s for anyone who wants power, choice, and control.\r\n\r\nMini Takeaway\r\n\r\nIf you finish this lesson understanding only one thing, let it be this:\r\nLinux isn’t hard — it just gives you more control than other systems. And you’re about to learn that control step by step.\r\n\r\nAssignment (super simple)\r\n\r\nNo installations yet. Just look this up and write down your observation:\r\n\r\nSearch “Linux distributions”\r\n\r\nLook at screenshots or homepages of at least 3 distros\r\n\r\nNote what vibe each distro gives you (beginner-friendly? minimalist? professional? etc.)\r\n\r\nThat’s it. Just getting a feel for the ecosystem.	/uploads/videos/4_1764224801.mp4	/uploads/files/4_1764224801.pdf	1	0	2025-11-27 06:26:41.150293	2025-11-27 14:26:02.970767
10	2	Intro	**Test Development**			1	0	2025-11-30 07:24:43.645859	2025-11-30 07:24:43.645859
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notifications (id, user_id, title, message, type, is_read, created_at) FROM stdin;
2	7	Course Completed	Congratulations! You have completed the course: 	completed	f	2025-11-27 11:16:02.162154
3	7	Course Completed	Congratulations! You have completed the course: Belajar Golang	completed	f	2025-11-27 11:27:10.889258
4	7	Course Completed	Congratulations! You have completed the course: Belajar Golang	completed	f	2025-11-27 12:49:33.619541
5	7	Course Completed	Congratulations! You have completed the course: Belajar Golang	completed	f	2025-11-27 12:55:15.837548
6	7	Course Completed	Congratulations! You have completed the course: Belajar Golang	completed	f	2025-11-27 12:56:51.362106
8	7	Course Completed	Congratulations! You have completed the course: Linux 101	completed	f	2025-11-27 13:45:46.123878
11	13	Account Created	Your account has been created by an admin.	completed	f	2025-11-30 15:58:21.335085
1	8	New Student Enrolled	A student has enrolled in your course: AWS for beginner	enrollment	t	2025-11-25 17:24:16.509209
7	8	New Student Enrolled	A student has enrolled in your course: Linux 101	enrollment	t	2025-11-27 13:45:07.166289
12	8	New Student Enrolled	A student has enrolled in your course: Linux 101	enrollment	t	2025-11-30 15:58:57.830396
14	14	Account Created	Your account has been created by an admin.	completed	f	2025-11-30 16:05:37.514004
15	6	User Created	You created user ID 14.	completed	f	2025-11-30 16:05:37.516298
13	6	User Deleted	You deleted user ID 11.	completed	t	2025-11-30 16:05:16.352205
\.


--
-- Data for Name: progress; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.progress (id, user_id, lesson_id, is_completed, completed_at) FROM stdin;
10	7	9	t	2025-11-27 13:45:46.114016
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, name, email, password, role, avatar, created_at, updated_at) FROM stdin;
8	Jessica	jess@smkn1garut.sch.id	$2a$10$MgnFUwxUoPyid6QpNfpGZ.elUqKdgJMeBacL6Z4rlNvQY2zyF3mhW	teacher	/uploads/avatars/8_1764132907.jpg	2025-11-25 09:24:39.269238	2025-11-26 11:55:07.95362
7	Student	joe@smkn1garut.sch.id	$2a$10$qpIwuyNtBZ9BBikEG.R8YObd.fZDmeJEXdLVs/Am8/No9TP8/u.HS	student	\N	2025-11-25 05:06:24.171086	2025-11-27 13:18:11.424405
6	Admin User	admin@test.com	$2a$10$27b981y1yAfNUt1cEOkyPuk5eeUtVsN4HKJQ.c3zB94vE2WJBrEGm	admin	\N	2025-11-25 02:46:36.833857	2025-11-30 14:21:45.206831
13	stud	stud@mail.com	$2a$10$YEV8joyw8f7r8nEZCavQ..PD3pubYWyan4Tz3Ys5cdAjBM7Dg4x/C	student	\N	2025-11-30 08:58:21.328758	2025-11-30 08:58:21.328758
14	Jane Doe	jane@doe.lol	$2a$10$LUmQTShKOvQVckuPE72CI.w27N1q/G8OqSQovCBbVqOsXRlB2O5rK	student	\N	2025-11-30 09:05:37.508768	2025-11-30 09:05:37.508768
\.


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.categories_id_seq', 4, true);


--
-- Name: courses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.courses_id_seq', 4, true);


--
-- Name: enrollments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.enrollments_id_seq', 4, true);


--
-- Name: lessons_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.lessons_id_seq', 10, true);


--
-- Name: notifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notifications_id_seq', 15, true);


--
-- Name: progress_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.progress_id_seq', 10, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 14, true);


--
-- Name: categories categories_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_name_key UNIQUE (name);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: courses courses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_pkey PRIMARY KEY (id);


--
-- Name: enrollments enrollments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.enrollments
    ADD CONSTRAINT enrollments_pkey PRIMARY KEY (id);


--
-- Name: enrollments enrollments_user_id_course_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.enrollments
    ADD CONSTRAINT enrollments_user_id_course_id_key UNIQUE (user_id, course_id);


--
-- Name: lessons lessons_course_id_order_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lessons
    ADD CONSTRAINT lessons_course_id_order_number_key UNIQUE (course_id, order_number);


--
-- Name: lessons lessons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lessons
    ADD CONSTRAINT lessons_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: progress progress_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.progress
    ADD CONSTRAINT progress_pkey PRIMARY KEY (id);


--
-- Name: progress progress_user_id_lesson_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.progress
    ADD CONSTRAINT progress_user_id_lesson_id_key UNIQUE (user_id, lesson_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_categories_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_categories_name ON public.categories USING btree (name);


--
-- Name: idx_courses_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_courses_category ON public.courses USING btree (category_id);


--
-- Name: idx_courses_published; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_courses_published ON public.courses USING btree (is_published);


--
-- Name: idx_courses_teacher; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_courses_teacher ON public.courses USING btree (teacher_id);


--
-- Name: idx_courses_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_courses_title ON public.courses USING btree (title);


--
-- Name: idx_enrollments_course; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_enrollments_course ON public.enrollments USING btree (course_id);


--
-- Name: idx_enrollments_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_enrollments_status ON public.enrollments USING btree (status);


--
-- Name: idx_enrollments_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_enrollments_user ON public.enrollments USING btree (user_id);


--
-- Name: idx_lessons_course; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_lessons_course ON public.lessons USING btree (course_id);


--
-- Name: idx_lessons_order; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_lessons_order ON public.lessons USING btree (course_id, order_number);


--
-- Name: idx_notifications_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notifications_created ON public.notifications USING btree (created_at DESC);


--
-- Name: idx_notifications_read; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notifications_read ON public.notifications USING btree (user_id, is_read);


--
-- Name: idx_notifications_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notifications_user ON public.notifications USING btree (user_id);


--
-- Name: idx_progress_lesson; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_progress_lesson ON public.progress USING btree (lesson_id);


--
-- Name: idx_progress_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_progress_user ON public.progress USING btree (user_id);


--
-- Name: idx_progress_user_lesson; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_progress_user_lesson ON public.progress USING btree (user_id, lesson_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_role ON public.users USING btree (role);


--
-- Name: course_details _RETURN; Type: RULE; Schema: public; Owner: postgres
--

CREATE OR REPLACE VIEW public.course_details AS
 SELECT c.id,
    c.title,
    c.description,
    c.thumbnail,
    c.is_published,
    cat.name AS category_name,
    u.name AS teacher_name,
    u.email AS teacher_email,
    count(DISTINCT l.id) AS total_lessons,
    count(DISTINCT e.id) AS total_students,
    c.created_at,
    c.updated_at
   FROM ((((public.courses c
     LEFT JOIN public.categories cat ON ((c.category_id = cat.id)))
     LEFT JOIN public.users u ON ((c.teacher_id = u.id)))
     LEFT JOIN public.lessons l ON ((c.id = l.course_id)))
     LEFT JOIN public.enrollments e ON (((c.id = e.course_id) AND (e.status = 'active'::public.enrollment_status))))
  GROUP BY c.id, cat.name, u.name, u.email;


--
-- Name: courses update_courses_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_courses_updated_at BEFORE UPDATE ON public.courses FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: lessons update_lessons_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_lessons_updated_at BEFORE UPDATE ON public.lessons FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: courses courses_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE SET NULL;


--
-- Name: courses courses_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: enrollments enrollments_course_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.enrollments
    ADD CONSTRAINT enrollments_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;


--
-- Name: enrollments enrollments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.enrollments
    ADD CONSTRAINT enrollments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: lessons lessons_course_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.lessons
    ADD CONSTRAINT lessons_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: progress progress_lesson_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.progress
    ADD CONSTRAINT progress_lesson_id_fkey FOREIGN KEY (lesson_id) REFERENCES public.lessons(id) ON DELETE CASCADE;


--
-- Name: progress progress_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.progress
    ADD CONSTRAINT progress_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict En0mRVaFoxH5CtcVf8wa2ADfiWdCX0jx4dIIMaq8bekEyZTXpNitEJJ3qSC43l6

