--
-- PostgreSQL database dump
--

-- Dumped from database version 17.7
-- Dumped by pg_dump version 17.0

-- Started on 2026-04-06 19:56:18

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
-- TOC entry 11 (class 2615 OID 16545)
-- Name: analytics_priv; Type: SCHEMA; Schema: -; Owner: schema_manager
--

CREATE SCHEMA analytics_priv;


ALTER SCHEMA analytics_priv OWNER TO schema_manager;

--
-- TOC entry 10 (class 2615 OID 16546)
-- Name: billing_secure; Type: SCHEMA; Schema: -; Owner: schema_manager
--

CREATE SCHEMA billing_secure;


ALTER SCHEMA billing_secure OWNER TO schema_manager;

--
-- TOC entry 9 (class 2615 OID 16544)
-- Name: core_data; Type: SCHEMA; Schema: -; Owner: schema_manager
--

CREATE SCHEMA core_data;


ALTER SCHEMA core_data OWNER TO schema_manager;

--
-- TOC entry 4 (class 3079 OID 16436)
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- TOC entry 3993 (class 0 OID 0)
-- Dependencies: 4
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- TOC entry 3 (class 3079 OID 16399)
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- TOC entry 3994 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- TOC entry 2 (class 3079 OID 16388)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 3995 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 1015 (class 1247 OID 17204)
-- Name: sub_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.sub_status AS ENUM (
    'free',
    'basic',
    'premium'
);


ALTER TYPE public.sub_status OWNER TO postgres;

--
-- TOC entry 1009 (class 1247 OID 17179)
-- Name: title_format; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.title_format AS ENUM (
    'movie',
    'series',
    'mini-series',
    'animation',
    'short',
    'tv_special'
);


ALTER TYPE public.title_format OWNER TO postgres;

--
-- TOC entry 1012 (class 1247 OID 17192)
-- Name: title_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.title_status AS ENUM (
    'released',
    'finished',
    'canceled',
    'in_production',
    'rumoredpilot'
);


ALTER TYPE public.title_status OWNER TO postgres;

--
-- TOC entry 1024 (class 1247 OID 17240)
-- Name: trivia_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.trivia_type AS ENUM (
    'fact',
    'blooper',
    'cameo',
    'callback',
    'production',
    'historical',
    'casting',
    'location',
    'deleted_scene'
);


ALTER TYPE public.trivia_type OWNER TO postgres;

--
-- TOC entry 1021 (class 1247 OID 17224)
-- Name: version_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.version_type AS ENUM (
    'basic',
    'directors_cut',
    'extended',
    'unrated',
    'pilot',
    'remastered',
    'black_and_white'
);


ALTER TYPE public.version_type OWNER TO postgres;

--
-- TOC entry 1018 (class 1247 OID 17212)
-- Name: watch_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.watch_status AS ENUM (
    'planned',
    'watching',
    'completed',
    'on_hold',
    'dropped'
);


ALTER TYPE public.watch_status OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 227 (class 1259 OID 16573)
-- Name: transaction_logs; Type: TABLE; Schema: analytics_priv; Owner: postgres
--

CREATE TABLE analytics_priv.transaction_logs (
    log_id integer NOT NULL,
    wallet_id integer NOT NULL,
    amount numeric(15,2) NOT NULL,
    tx_type character varying(10),
    executed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT transaction_logs_tx_type_check CHECK (((tx_type)::text = ANY ((ARRAY['deposit'::character varying, 'withdraw'::character varying])::text[])))
);


ALTER TABLE analytics_priv.transaction_logs OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16572)
-- Name: transaction_logs_log_id_seq; Type: SEQUENCE; Schema: analytics_priv; Owner: postgres
--

CREATE SEQUENCE analytics_priv.transaction_logs_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE analytics_priv.transaction_logs_log_id_seq OWNER TO postgres;

--
-- TOC entry 3997 (class 0 OID 0)
-- Dependencies: 226
-- Name: transaction_logs_log_id_seq; Type: SEQUENCE OWNED BY; Schema: analytics_priv; Owner: postgres
--

ALTER SEQUENCE analytics_priv.transaction_logs_log_id_seq OWNED BY analytics_priv.transaction_logs.log_id;


--
-- TOC entry 225 (class 1259 OID 16559)
-- Name: wallets; Type: TABLE; Schema: billing_secure; Owner: postgres
--

CREATE TABLE billing_secure.wallets (
    wallet_id integer NOT NULL,
    owner_id uuid NOT NULL,
    balance numeric(15,2) DEFAULT 0.00,
    CONSTRAINT wallets_balance_check CHECK ((balance >= (0)::numeric))
);


ALTER TABLE billing_secure.wallets OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16558)
-- Name: wallets_wallet_id_seq; Type: SEQUENCE; Schema: billing_secure; Owner: postgres
--

CREATE SEQUENCE billing_secure.wallets_wallet_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE billing_secure.wallets_wallet_id_seq OWNER TO postgres;

--
-- TOC entry 3999 (class 0 OID 0)
-- Dependencies: 224
-- Name: wallets_wallet_id_seq; Type: SEQUENCE OWNED BY; Schema: billing_secure; Owner: postgres
--

ALTER SEQUENCE billing_secure.wallets_wallet_id_seq OWNED BY billing_secure.wallets.wallet_id;


--
-- TOC entry 223 (class 1259 OID 16547)
-- Name: users; Type: TABLE; Schema: core_data; Owner: postgres
--

CREATE TABLE core_data.users (
    user_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    username public.citext NOT NULL,
    pass_hash text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE core_data.users OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 16586)
-- Name: v_rich_users_report; Type: VIEW; Schema: core_data; Owner: postgres
--

CREATE VIEW core_data.v_rich_users_report AS
 SELECT u.username,
    w.balance,
    t.amount AS last_tx_amount
   FROM ((core_data.users u
     JOIN billing_secure.wallets w ON ((u.user_id = w.owner_id)))
     LEFT JOIN analytics_priv.transaction_logs t ON ((w.wallet_id = t.wallet_id)))
  WHERE (w.balance > (1000)::numeric);


ALTER VIEW core_data.v_rich_users_report OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 17260)
-- Name: account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account (
    account_id bigint NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    is_verified boolean DEFAULT false,
    is_banned boolean DEFAULT false,
    subscription_status public.sub_status DEFAULT 'free'::public.sub_status,
    subscription_end_date timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    last_login_at timestamp without time zone
);


ALTER TABLE public.account OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 17259)
-- Name: account_account_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.account ALTER COLUMN account_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.account_account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 242 (class 1259 OID 17371)
-- Name: award; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.award (
    award_id integer NOT NULL,
    name text NOT NULL,
    name_orig text NOT NULL,
    city text NOT NULL,
    country_id smallint,
    description text
);


ALTER TABLE public.award OWNER TO postgres;

--
-- TOC entry 241 (class 1259 OID 17370)
-- Name: award_award_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.award ALTER COLUMN award_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.award_award_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 245 (class 1259 OID 17397)
-- Name: award_category; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.award_category (
    award_id integer NOT NULL,
    category_id integer NOT NULL
);


ALTER TABLE public.award_category OWNER TO postgres;

--
-- TOC entry 244 (class 1259 OID 17388)
-- Name: category; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.category (
    category_id integer NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.category OWNER TO postgres;

--
-- TOC entry 243 (class 1259 OID 17387)
-- Name: category_category_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.category ALTER COLUMN category_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.category_category_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 269 (class 1259 OID 17630)
-- Name: content_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.content_info (
    content_id bigint NOT NULL,
    title_id bigint,
    episode_id bigint,
    version_type public.version_type DEFAULT 'basic'::public.version_type NOT NULL,
    note text,
    file_path text NOT NULL,
    quality character varying(5) NOT NULL,
    v_codec character varying(5) DEFAULT 'h264'::character varying,
    a_codec character varying(5) DEFAULT 'aac'::character varying,
    is_hdr boolean DEFAULT false,
    size_bytes bigint,
    duration_seconds integer,
    audio_languages character(2)[] DEFAULT '{}'::bpchar[],
    subtitle_languages character(2)[] DEFAULT '{}'::bpchar[],
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT check_content_link CHECK ((((title_id IS NOT NULL) AND (episode_id IS NULL)) OR ((title_id IS NULL) AND (episode_id IS NOT NULL)))),
    CONSTRAINT content_info_duration_seconds_check CHECK ((duration_seconds > 0)),
    CONSTRAINT content_info_size_bytes_check CHECK ((size_bytes > 0))
);


ALTER TABLE public.content_info OWNER TO postgres;

--
-- TOC entry 268 (class 1259 OID 17629)
-- Name: content_info_content_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.content_info ALTER COLUMN content_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.content_info_content_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 239 (class 1259 OID 17344)
-- Name: country; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.country (
    country_id smallint NOT NULL,
    name text NOT NULL,
    iso_code character(2)
);


ALTER TABLE public.country OWNER TO postgres;

--
-- TOC entry 238 (class 1259 OID 17343)
-- Name: country_country_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.country ALTER COLUMN country_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.country_country_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 249 (class 1259 OID 17424)
-- Name: department; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.department (
    department_id smallint NOT NULL,
    name_ru text NOT NULL,
    name_en text NOT NULL
);


ALTER TABLE public.department OWNER TO postgres;

--
-- TOC entry 248 (class 1259 OID 17423)
-- Name: department_department_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.department ALTER COLUMN department_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.department_department_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 259 (class 1259 OID 17519)
-- Name: episode; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.episode (
    episode_id bigint NOT NULL,
    season_id bigint NOT NULL,
    episode_number smallint NOT NULL,
    title_ru text,
    title_origin text,
    description text,
    release_date date,
    duration smallint,
    rating numeric(3,1),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT episode_duration_check CHECK ((duration > 0)),
    CONSTRAINT episode_episode_number_check CHECK ((episode_number > 0)),
    CONSTRAINT episode_rating_check CHECK (((rating >= (0)::numeric) AND (rating <= (10)::numeric)))
);


ALTER TABLE public.episode OWNER TO postgres;

--
-- TOC entry 258 (class 1259 OID 17518)
-- Name: episode_episode_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.episode ALTER COLUMN episode_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.episode_episode_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 265 (class 1259 OID 17586)
-- Name: episode_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.episode_tag (
    episode_id bigint NOT NULL,
    tag_id integer NOT NULL,
    user_id bigint NOT NULL,
    is_private boolean DEFAULT false,
    is_spoiler boolean DEFAULT false
);


ALTER TABLE public.episode_tag OWNER TO postgres;

--
-- TOC entry 236 (class 1259 OID 17317)
-- Name: genre; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.genre (
    genre_id integer NOT NULL,
    name text NOT NULL,
    slug text
);


ALTER TABLE public.genre OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 17316)
-- Name: genre_genre_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.genre ALTER COLUMN genre_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.genre_genre_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 247 (class 1259 OID 17413)
-- Name: person; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.person (
    person_id bigint NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    original_name text,
    birth_date date NOT NULL,
    death_date date,
    photo_url text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT person_check CHECK ((death_date > birth_date))
);


ALTER TABLE public.person OWNER TO postgres;

--
-- TOC entry 246 (class 1259 OID 17412)
-- Name: person_person_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.person ALTER COLUMN person_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.person_person_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 251 (class 1259 OID 17436)
-- Name: profession; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.profession (
    profession_id integer NOT NULL,
    department_id smallint,
    name_ru text NOT NULL,
    name_en text NOT NULL
);


ALTER TABLE public.profession OWNER TO postgres;

--
-- TOC entry 250 (class 1259 OID 17435)
-- Name: profession_profession_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.profession ALTER COLUMN profession_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.profession_profession_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 232 (class 1259 OID 17274)
-- Name: profile; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.profile (
    profile_id bigint NOT NULL,
    account_id bigint NOT NULL,
    name text NOT NULL,
    avatar_url text,
    is_kid boolean DEFAULT false,
    age_limit smallint DEFAULT 18 NOT NULL,
    pin_code character(4),
    language character varying(5) DEFAULT 'ru'::character varying,
    is_autoplay_next boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.profile OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 17273)
-- Name: profile_profile_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.profile ALTER COLUMN profile_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.profile_profile_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 257 (class 1259 OID 17500)
-- Name: season; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.season (
    season_id bigint NOT NULL,
    title_id bigint NOT NULL,
    season_number smallint NOT NULL,
    title_ru text,
    title_origin text,
    release_date date,
    end_date date,
    description text,
    poster_path text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT check_season_dates CHECK ((end_date >= release_date)),
    CONSTRAINT season_season_number_check CHECK ((season_number > 0))
);


ALTER TABLE public.season OWNER TO postgres;

--
-- TOC entry 256 (class 1259 OID 17499)
-- Name: season_season_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.season ALTER COLUMN season_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.season_season_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 263 (class 1259 OID 17550)
-- Name: tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tag (
    tag_id bigint NOT NULL,
    category_id integer NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    description text,
    is_system boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tag OWNER TO postgres;

--
-- TOC entry 261 (class 1259 OID 17539)
-- Name: tag_category; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tag_category (
    category_id integer NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    is_visible boolean DEFAULT true
);


ALTER TABLE public.tag_category OWNER TO postgres;

--
-- TOC entry 260 (class 1259 OID 17538)
-- Name: tag_category_category_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.tag_category ALTER COLUMN category_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.tag_category_category_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 262 (class 1259 OID 17549)
-- Name: tag_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.tag ALTER COLUMN tag_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.tag_tag_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 234 (class 1259 OID 17294)
-- Name: title; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.title (
    title_id bigint NOT NULL,
    title_ru text,
    title_origin text NOT NULL,
    release_year smallint NOT NULL,
    end_year smallint,
    format public.title_format NOT NULL,
    rating_imdb numeric(3,1),
    rating_kinopoisk numeric(3,1),
    metascore smallint,
    status public.title_status DEFAULT 'released'::public.title_status,
    description text,
    poster_path text,
    budget numeric(14,2),
    world_fees numeric(14,2),
    age_rating smallint NOT NULL,
    duration smallint,
    is_published boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT title_age_rating_check CHECK ((age_rating = ANY (ARRAY[0, 6, 12, 16, 18]))),
    CONSTRAINT title_budget_check CHECK ((budget >= (0)::numeric)),
    CONSTRAINT title_check CHECK ((end_year >= release_year)),
    CONSTRAINT title_duration_check CHECK ((duration > 0)),
    CONSTRAINT title_metascore_check CHECK (((metascore >= 0) AND (metascore <= 100))),
    CONSTRAINT title_rating_imdb_check CHECK (((rating_imdb >= (0)::numeric) AND (rating_imdb <= (10)::numeric))),
    CONSTRAINT title_rating_kinopoisk_check CHECK (((rating_kinopoisk >= (0)::numeric) AND (rating_kinopoisk <= (10)::numeric))),
    CONSTRAINT title_release_year_check CHECK (((release_year >= 1888) AND (release_year <= 2200))),
    CONSTRAINT title_world_fees_check CHECK ((world_fees >= (0)::numeric))
);


ALTER TABLE public.title OWNER TO postgres;

--
-- TOC entry 255 (class 1259 OID 17476)
-- Name: title_award; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.title_award (
    title_award_id bigint NOT NULL,
    award_id integer NOT NULL,
    category_id integer NOT NULL,
    title_id bigint NOT NULL,
    person_id bigint,
    year smallint NOT NULL,
    is_winner boolean DEFAULT false
);


ALTER TABLE public.title_award OWNER TO postgres;

--
-- TOC entry 254 (class 1259 OID 17475)
-- Name: title_award_title_award_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.title_award ALTER COLUMN title_award_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.title_award_title_award_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 240 (class 1259 OID 17355)
-- Name: title_country; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.title_country (
    title_id bigint NOT NULL,
    country_id smallint NOT NULL
);


ALTER TABLE public.title_country OWNER TO postgres;

--
-- TOC entry 237 (class 1259 OID 17328)
-- Name: title_genre; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.title_genre (
    title_id bigint NOT NULL,
    genre_id integer NOT NULL
);


ALTER TABLE public.title_genre OWNER TO postgres;

--
-- TOC entry 253 (class 1259 OID 17451)
-- Name: title_person; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.title_person (
    title_person_id bigint NOT NULL,
    title_id bigint,
    person_id bigint,
    profession_id integer,
    character_name text
);


ALTER TABLE public.title_person OWNER TO postgres;

--
-- TOC entry 252 (class 1259 OID 17450)
-- Name: title_person_title_person_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.title_person ALTER COLUMN title_person_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.title_person_title_person_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 264 (class 1259 OID 17568)
-- Name: title_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.title_tag (
    title_id bigint NOT NULL,
    tag_id integer NOT NULL,
    user_id bigint NOT NULL,
    is_private boolean DEFAULT false,
    is_spoiler boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.title_tag OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 17293)
-- Name: title_title_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.title ALTER COLUMN title_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.title_title_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 273 (class 1259 OID 17687)
-- Name: trivia; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.trivia (
    trivia_id bigint NOT NULL,
    content text NOT NULL,
    type public.trivia_type DEFAULT 'fact'::public.trivia_type NOT NULL,
    sort_rank integer DEFAULT 0,
    is_published boolean DEFAULT true,
    is_spoiler boolean DEFAULT false,
    at_timestamp_sec integer,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.trivia OWNER TO postgres;

--
-- TOC entry 275 (class 1259 OID 17701)
-- Name: trivia_link; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.trivia_link (
    id bigint NOT NULL,
    trivia_id bigint NOT NULL,
    title_id bigint,
    person_id bigint,
    episode_id bigint,
    CONSTRAINT trivia_link_target_check CHECK ((((((title_id IS NOT NULL))::integer + ((person_id IS NOT NULL))::integer) + ((episode_id IS NOT NULL))::integer) >= 1))
);


ALTER TABLE public.trivia_link OWNER TO postgres;

--
-- TOC entry 274 (class 1259 OID 17700)
-- Name: trivia_link_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.trivia_link ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.trivia_link_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 272 (class 1259 OID 17686)
-- Name: trivia_trivia_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.trivia ALTER COLUMN trivia_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.trivia_trivia_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 267 (class 1259 OID 17604)
-- Name: user_collection; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_collection (
    collection_id bigint NOT NULL,
    profile_id bigint NOT NULL,
    title_id bigint NOT NULL,
    status public.watch_status DEFAULT 'planned'::public.watch_status NOT NULL,
    score numeric(3,1),
    review_text text,
    review_title text,
    is_spoiler boolean DEFAULT false,
    is_gold boolean DEFAULT false,
    is_liked boolean,
    watched_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_collection_score_check CHECK (((score >= (1)::numeric) AND (score <= (10)::numeric)))
);


ALTER TABLE public.user_collection OWNER TO postgres;

--
-- TOC entry 266 (class 1259 OID 17603)
-- Name: user_collection_collection_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.user_collection ALTER COLUMN collection_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.user_collection_collection_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 271 (class 1259 OID 17660)
-- Name: watch_progress; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.watch_progress (
    progress_id bigint NOT NULL,
    profile_id bigint NOT NULL,
    title_id bigint,
    episode_id bigint,
    current_time_sec integer DEFAULT 0 NOT NULL,
    is_completed boolean DEFAULT false,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT check_progress_target CHECK ((((title_id IS NOT NULL) AND (episode_id IS NULL)) OR ((title_id IS NULL) AND (episode_id IS NOT NULL))))
);


ALTER TABLE public.watch_progress OWNER TO postgres;

--
-- TOC entry 270 (class 1259 OID 17659)
-- Name: watch_progress_progress_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.watch_progress ALTER COLUMN progress_id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.watch_progress_progress_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 3572 (class 2604 OID 16576)
-- Name: transaction_logs log_id; Type: DEFAULT; Schema: analytics_priv; Owner: postgres
--

ALTER TABLE ONLY analytics_priv.transaction_logs ALTER COLUMN log_id SET DEFAULT nextval('analytics_priv.transaction_logs_log_id_seq'::regclass);


--
-- TOC entry 3570 (class 2604 OID 16562)
-- Name: wallets wallet_id; Type: DEFAULT; Schema: billing_secure; Owner: postgres
--

ALTER TABLE ONLY billing_secure.wallets ALTER COLUMN wallet_id SET DEFAULT nextval('billing_secure.wallets_wallet_id_seq'::regclass);


--
-- TOC entry 3937 (class 0 OID 16573)
-- Dependencies: 227
-- Data for Name: transaction_logs; Type: TABLE DATA; Schema: analytics_priv; Owner: postgres
--

COPY analytics_priv.transaction_logs (log_id, wallet_id, amount, tx_type, executed_at) FROM stdin;
1	2	500.00	deposit	2026-04-05 18:59:13.749876
2	1	100.00	withdraw	2026-04-05 18:59:13.749876
\.


--
-- TOC entry 3935 (class 0 OID 16559)
-- Dependencies: 225
-- Data for Name: wallets; Type: TABLE DATA; Schema: billing_secure; Owner: postgres
--

COPY billing_secure.wallets (wallet_id, owner_id, balance) FROM stdin;
1	9c64614e-fe48-4b53-980a-c847461c4e46	1500.00
2	7d8d1f1f-10bd-46ff-b5e8-a9007425e1c1	99999.50
3	f1c01532-5d8f-48d5-8f3c-6fb95040b452	0.00
\.


--
-- TOC entry 3933 (class 0 OID 16547)
-- Dependencies: 223
-- Data for Name: users; Type: TABLE DATA; Schema: core_data; Owner: postgres
--

COPY core_data.users (user_id, username, pass_hash, created_at) FROM stdin;
9c64614e-fe48-4b53-980a-c847461c4e46	alex_dev	$2a$06$I14kbuZUai9e/DdKt4kloO6aojEIdo44r.jNzyR43IS8efKaqRfpq	2026-04-05 18:59:13.733587
7d8d1f1f-10bd-46ff-b5e8-a9007425e1c1	murad_admin	$2a$06$sfu/ayWIHbe/7FPySbrZX.hXOeTVvKAtnxbcq74IwCwLFmICuwYK2	2026-04-05 18:59:13.733587
f1c01532-5d8f-48d5-8f3c-6fb95040b452	guest_user	$2a$06$VZBVsKF/lslv2lovNU0GM.K31fGRl4MjDK.gYurJD0j0M93kOv2Dq	2026-04-05 18:59:13.733587
\.


--
-- TOC entry 3939 (class 0 OID 17260)
-- Dependencies: 230
-- Data for Name: account; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.account (account_id, email, password_hash, is_verified, is_banned, subscription_status, subscription_end_date, created_at, last_login_at) FROM stdin;
\.


--
-- TOC entry 3951 (class 0 OID 17371)
-- Dependencies: 242
-- Data for Name: award; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.award (award_id, name, name_orig, city, country_id, description) FROM stdin;
\.


--
-- TOC entry 3954 (class 0 OID 17397)
-- Dependencies: 245
-- Data for Name: award_category; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.award_category (award_id, category_id) FROM stdin;
\.


--
-- TOC entry 3953 (class 0 OID 17388)
-- Dependencies: 244
-- Data for Name: category; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.category (category_id, name) FROM stdin;
\.


--
-- TOC entry 3978 (class 0 OID 17630)
-- Dependencies: 269
-- Data for Name: content_info; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.content_info (content_id, title_id, episode_id, version_type, note, file_path, quality, v_codec, a_codec, is_hdr, size_bytes, duration_seconds, audio_languages, subtitle_languages, created_at) FROM stdin;
\.


--
-- TOC entry 3948 (class 0 OID 17344)
-- Dependencies: 239
-- Data for Name: country; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.country (country_id, name, iso_code) FROM stdin;
\.


--
-- TOC entry 3958 (class 0 OID 17424)
-- Dependencies: 249
-- Data for Name: department; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.department (department_id, name_ru, name_en) FROM stdin;
\.


--
-- TOC entry 3968 (class 0 OID 17519)
-- Dependencies: 259
-- Data for Name: episode; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.episode (episode_id, season_id, episode_number, title_ru, title_origin, description, release_date, duration, rating, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 3974 (class 0 OID 17586)
-- Dependencies: 265
-- Data for Name: episode_tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.episode_tag (episode_id, tag_id, user_id, is_private, is_spoiler) FROM stdin;
\.


--
-- TOC entry 3945 (class 0 OID 17317)
-- Dependencies: 236
-- Data for Name: genre; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.genre (genre_id, name, slug) FROM stdin;
\.


--
-- TOC entry 3956 (class 0 OID 17413)
-- Dependencies: 247
-- Data for Name: person; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.person (person_id, first_name, last_name, original_name, birth_date, death_date, photo_url, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 3960 (class 0 OID 17436)
-- Dependencies: 251
-- Data for Name: profession; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.profession (profession_id, department_id, name_ru, name_en) FROM stdin;
\.


--
-- TOC entry 3941 (class 0 OID 17274)
-- Dependencies: 232
-- Data for Name: profile; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.profile (profile_id, account_id, name, avatar_url, is_kid, age_limit, pin_code, language, is_autoplay_next, created_at) FROM stdin;
\.


--
-- TOC entry 3966 (class 0 OID 17500)
-- Dependencies: 257
-- Data for Name: season; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.season (season_id, title_id, season_number, title_ru, title_origin, release_date, end_date, description, poster_path, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 3972 (class 0 OID 17550)
-- Dependencies: 263
-- Data for Name: tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tag (tag_id, category_id, name, slug, description, is_system, created_at) FROM stdin;
\.


--
-- TOC entry 3970 (class 0 OID 17539)
-- Dependencies: 261
-- Data for Name: tag_category; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tag_category (category_id, slug, name, is_visible) FROM stdin;
\.


--
-- TOC entry 3943 (class 0 OID 17294)
-- Dependencies: 234
-- Data for Name: title; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.title (title_id, title_ru, title_origin, release_year, end_year, format, rating_imdb, rating_kinopoisk, metascore, status, description, poster_path, budget, world_fees, age_rating, duration, is_published, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 3964 (class 0 OID 17476)
-- Dependencies: 255
-- Data for Name: title_award; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.title_award (title_award_id, award_id, category_id, title_id, person_id, year, is_winner) FROM stdin;
\.


--
-- TOC entry 3949 (class 0 OID 17355)
-- Dependencies: 240
-- Data for Name: title_country; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.title_country (title_id, country_id) FROM stdin;
\.


--
-- TOC entry 3946 (class 0 OID 17328)
-- Dependencies: 237
-- Data for Name: title_genre; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.title_genre (title_id, genre_id) FROM stdin;
\.


--
-- TOC entry 3962 (class 0 OID 17451)
-- Dependencies: 253
-- Data for Name: title_person; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.title_person (title_person_id, title_id, person_id, profession_id, character_name) FROM stdin;
\.


--
-- TOC entry 3973 (class 0 OID 17568)
-- Dependencies: 264
-- Data for Name: title_tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.title_tag (title_id, tag_id, user_id, is_private, is_spoiler, created_at) FROM stdin;
\.


--
-- TOC entry 3982 (class 0 OID 17687)
-- Dependencies: 273
-- Data for Name: trivia; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.trivia (trivia_id, content, type, sort_rank, is_published, is_spoiler, at_timestamp_sec, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 3984 (class 0 OID 17701)
-- Dependencies: 275
-- Data for Name: trivia_link; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.trivia_link (id, trivia_id, title_id, person_id, episode_id) FROM stdin;
\.


--
-- TOC entry 3976 (class 0 OID 17604)
-- Dependencies: 267
-- Data for Name: user_collection; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_collection (collection_id, profile_id, title_id, status, score, review_text, review_title, is_spoiler, is_gold, is_liked, watched_at, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 3980 (class 0 OID 17660)
-- Dependencies: 271
-- Data for Name: watch_progress; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.watch_progress (progress_id, profile_id, title_id, episode_id, current_time_sec, is_completed, updated_at) FROM stdin;
\.


--
-- TOC entry 4002 (class 0 OID 0)
-- Dependencies: 226
-- Name: transaction_logs_log_id_seq; Type: SEQUENCE SET; Schema: analytics_priv; Owner: postgres
--

SELECT pg_catalog.setval('analytics_priv.transaction_logs_log_id_seq', 2, true);


--
-- TOC entry 4003 (class 0 OID 0)
-- Dependencies: 224
-- Name: wallets_wallet_id_seq; Type: SEQUENCE SET; Schema: billing_secure; Owner: postgres
--

SELECT pg_catalog.setval('billing_secure.wallets_wallet_id_seq', 3, true);


--
-- TOC entry 4004 (class 0 OID 0)
-- Dependencies: 229
-- Name: account_account_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.account_account_id_seq', 1, false);


--
-- TOC entry 4005 (class 0 OID 0)
-- Dependencies: 241
-- Name: award_award_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.award_award_id_seq', 1, false);


--
-- TOC entry 4006 (class 0 OID 0)
-- Dependencies: 243
-- Name: category_category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.category_category_id_seq', 1, false);


--
-- TOC entry 4007 (class 0 OID 0)
-- Dependencies: 268
-- Name: content_info_content_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.content_info_content_id_seq', 1, false);


--
-- TOC entry 4008 (class 0 OID 0)
-- Dependencies: 238
-- Name: country_country_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.country_country_id_seq', 1, false);


--
-- TOC entry 4009 (class 0 OID 0)
-- Dependencies: 248
-- Name: department_department_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.department_department_id_seq', 1, false);


--
-- TOC entry 4010 (class 0 OID 0)
-- Dependencies: 258
-- Name: episode_episode_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.episode_episode_id_seq', 1, false);


--
-- TOC entry 4011 (class 0 OID 0)
-- Dependencies: 235
-- Name: genre_genre_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.genre_genre_id_seq', 1, false);


--
-- TOC entry 4012 (class 0 OID 0)
-- Dependencies: 246
-- Name: person_person_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.person_person_id_seq', 1, false);


--
-- TOC entry 4013 (class 0 OID 0)
-- Dependencies: 250
-- Name: profession_profession_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.profession_profession_id_seq', 1, false);


--
-- TOC entry 4014 (class 0 OID 0)
-- Dependencies: 231
-- Name: profile_profile_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.profile_profile_id_seq', 1, false);


--
-- TOC entry 4015 (class 0 OID 0)
-- Dependencies: 256
-- Name: season_season_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.season_season_id_seq', 1, false);


--
-- TOC entry 4016 (class 0 OID 0)
-- Dependencies: 260
-- Name: tag_category_category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tag_category_category_id_seq', 1, false);


--
-- TOC entry 4017 (class 0 OID 0)
-- Dependencies: 262
-- Name: tag_tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tag_tag_id_seq', 1, false);


--
-- TOC entry 4018 (class 0 OID 0)
-- Dependencies: 254
-- Name: title_award_title_award_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.title_award_title_award_id_seq', 1, false);


--
-- TOC entry 4019 (class 0 OID 0)
-- Dependencies: 252
-- Name: title_person_title_person_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.title_person_title_person_id_seq', 1, false);


--
-- TOC entry 4020 (class 0 OID 0)
-- Dependencies: 233
-- Name: title_title_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.title_title_id_seq', 1, false);


--
-- TOC entry 4021 (class 0 OID 0)
-- Dependencies: 274
-- Name: trivia_link_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.trivia_link_id_seq', 1, false);


--
-- TOC entry 4022 (class 0 OID 0)
-- Dependencies: 272
-- Name: trivia_trivia_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.trivia_trivia_id_seq', 1, false);


--
-- TOC entry 4023 (class 0 OID 0)
-- Dependencies: 266
-- Name: user_collection_collection_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_collection_collection_id_seq', 1, false);


--
-- TOC entry 4024 (class 0 OID 0)
-- Dependencies: 270
-- Name: watch_progress_progress_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.watch_progress_progress_id_seq', 1, false);


--
-- TOC entry 3653 (class 2606 OID 16580)
-- Name: transaction_logs transaction_logs_pkey; Type: CONSTRAINT; Schema: analytics_priv; Owner: postgres
--

ALTER TABLE ONLY analytics_priv.transaction_logs
    ADD CONSTRAINT transaction_logs_pkey PRIMARY KEY (log_id);


--
-- TOC entry 3651 (class 2606 OID 16566)
-- Name: wallets wallets_pkey; Type: CONSTRAINT; Schema: billing_secure; Owner: postgres
--

ALTER TABLE ONLY billing_secure.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (wallet_id);


--
-- TOC entry 3647 (class 2606 OID 16555)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: core_data; Owner: postgres
--

ALTER TABLE ONLY core_data.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3649 (class 2606 OID 16557)
-- Name: users users_username_key; Type: CONSTRAINT; Schema: core_data; Owner: postgres
--

ALTER TABLE ONLY core_data.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- TOC entry 3655 (class 2606 OID 17272)
-- Name: account account_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_email_key UNIQUE (email);


--
-- TOC entry 3657 (class 2606 OID 17270)
-- Name: account account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_pkey PRIMARY KEY (account_id);


--
-- TOC entry 3693 (class 2606 OID 17401)
-- Name: award_category award_category_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award_category
    ADD CONSTRAINT award_category_pkey PRIMARY KEY (award_id, category_id);


--
-- TOC entry 3683 (class 2606 OID 17379)
-- Name: award award_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_name_key UNIQUE (name);


--
-- TOC entry 3685 (class 2606 OID 17381)
-- Name: award award_name_orig_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_name_orig_key UNIQUE (name_orig);


--
-- TOC entry 3687 (class 2606 OID 17377)
-- Name: award award_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_pkey PRIMARY KEY (award_id);


--
-- TOC entry 3689 (class 2606 OID 17396)
-- Name: category category_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_name_key UNIQUE (name);


--
-- TOC entry 3691 (class 2606 OID 17394)
-- Name: category category_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (category_id);


--
-- TOC entry 3741 (class 2606 OID 17646)
-- Name: content_info content_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.content_info
    ADD CONSTRAINT content_info_pkey PRIMARY KEY (content_id);


--
-- TOC entry 3675 (class 2606 OID 17354)
-- Name: country country_iso_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.country
    ADD CONSTRAINT country_iso_code_key UNIQUE (iso_code);


--
-- TOC entry 3677 (class 2606 OID 17352)
-- Name: country country_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.country
    ADD CONSTRAINT country_name_key UNIQUE (name);


--
-- TOC entry 3679 (class 2606 OID 17350)
-- Name: country country_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.country
    ADD CONSTRAINT country_pkey PRIMARY KEY (country_id);


--
-- TOC entry 3697 (class 2606 OID 17434)
-- Name: department department_name_en_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT department_name_en_key UNIQUE (name_en);


--
-- TOC entry 3699 (class 2606 OID 17432)
-- Name: department department_name_ru_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT department_name_ru_key UNIQUE (name_ru);


--
-- TOC entry 3701 (class 2606 OID 17430)
-- Name: department department_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT department_pkey PRIMARY KEY (department_id);


--
-- TOC entry 3719 (class 2606 OID 17530)
-- Name: episode episode_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.episode
    ADD CONSTRAINT episode_pkey PRIMARY KEY (episode_id);


--
-- TOC entry 3735 (class 2606 OID 17592)
-- Name: episode_tag episode_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.episode_tag
    ADD CONSTRAINT episode_tag_pkey PRIMARY KEY (episode_id, tag_id, user_id);


--
-- TOC entry 3667 (class 2606 OID 17325)
-- Name: genre genre_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre
    ADD CONSTRAINT genre_name_key UNIQUE (name);


--
-- TOC entry 3669 (class 2606 OID 17323)
-- Name: genre genre_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre
    ADD CONSTRAINT genre_pkey PRIMARY KEY (genre_id);


--
-- TOC entry 3671 (class 2606 OID 17327)
-- Name: genre genre_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre
    ADD CONSTRAINT genre_slug_key UNIQUE (slug);


--
-- TOC entry 3695 (class 2606 OID 17422)
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (person_id);


--
-- TOC entry 3703 (class 2606 OID 17442)
-- Name: profession profession_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profession
    ADD CONSTRAINT profession_pkey PRIMARY KEY (profession_id);


--
-- TOC entry 3659 (class 2606 OID 17285)
-- Name: profile profile_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_pkey PRIMARY KEY (profile_id);


--
-- TOC entry 3715 (class 2606 OID 17510)
-- Name: season season_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.season
    ADD CONSTRAINT season_pkey PRIMARY KEY (season_id);


--
-- TOC entry 3723 (class 2606 OID 17546)
-- Name: tag_category tag_category_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_category
    ADD CONSTRAINT tag_category_pkey PRIMARY KEY (category_id);


--
-- TOC entry 3725 (class 2606 OID 17548)
-- Name: tag_category tag_category_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag_category
    ADD CONSTRAINT tag_category_slug_key UNIQUE (slug);


--
-- TOC entry 3727 (class 2606 OID 17560)
-- Name: tag tag_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_name_key UNIQUE (name);


--
-- TOC entry 3729 (class 2606 OID 17558)
-- Name: tag tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_pkey PRIMARY KEY (tag_id);


--
-- TOC entry 3731 (class 2606 OID 17562)
-- Name: tag tag_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_slug_key UNIQUE (slug);


--
-- TOC entry 3711 (class 2606 OID 17481)
-- Name: title_award title_award_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_award
    ADD CONSTRAINT title_award_pkey PRIMARY KEY (title_award_id);


--
-- TOC entry 3681 (class 2606 OID 17359)
-- Name: title_country title_country_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_country
    ADD CONSTRAINT title_country_pkey PRIMARY KEY (title_id, country_id);


--
-- TOC entry 3673 (class 2606 OID 17332)
-- Name: title_genre title_genre_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_genre
    ADD CONSTRAINT title_genre_pkey PRIMARY KEY (title_id, genre_id);


--
-- TOC entry 3707 (class 2606 OID 17457)
-- Name: title_person title_person_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_person
    ADD CONSTRAINT title_person_pkey PRIMARY KEY (title_person_id);


--
-- TOC entry 3663 (class 2606 OID 17313)
-- Name: title title_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title
    ADD CONSTRAINT title_pkey PRIMARY KEY (title_id);


--
-- TOC entry 3733 (class 2606 OID 17575)
-- Name: title_tag title_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_tag
    ADD CONSTRAINT title_tag_pkey PRIMARY KEY (title_id, tag_id, user_id);


--
-- TOC entry 3751 (class 2606 OID 17706)
-- Name: trivia_link trivia_link_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.trivia_link
    ADD CONSTRAINT trivia_link_pkey PRIMARY KEY (id);


--
-- TOC entry 3749 (class 2606 OID 17699)
-- Name: trivia trivia_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.trivia
    ADD CONSTRAINT trivia_pkey PRIMARY KEY (trivia_id);


--
-- TOC entry 3713 (class 2606 OID 17483)
-- Name: title_award unique_award_nomination; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_award
    ADD CONSTRAINT unique_award_nomination UNIQUE (award_id, category_id, title_id, person_id, year);


--
-- TOC entry 3721 (class 2606 OID 17532)
-- Name: episode unique_episode_per_season; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.episode
    ADD CONSTRAINT unique_episode_per_season UNIQUE (season_id, episode_number);


--
-- TOC entry 3743 (class 2606 OID 17648)
-- Name: content_info unique_file_url; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.content_info
    ADD CONSTRAINT unique_file_url UNIQUE (file_path);


--
-- TOC entry 3705 (class 2606 OID 17444)
-- Name: profession unique_profession_name; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profession
    ADD CONSTRAINT unique_profession_name UNIQUE (department_id, name_en);


--
-- TOC entry 3661 (class 2606 OID 17287)
-- Name: profile unique_profile_name_per_account; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT unique_profile_name_per_account UNIQUE (account_id, name);


--
-- TOC entry 3737 (class 2606 OID 17618)
-- Name: user_collection unique_profile_title_collection; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_collection
    ADD CONSTRAINT unique_profile_title_collection UNIQUE (profile_id, title_id);


--
-- TOC entry 3745 (class 2606 OID 17670)
-- Name: watch_progress unique_progress_per_item; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_progress
    ADD CONSTRAINT unique_progress_per_item UNIQUE (profile_id, title_id, episode_id);


--
-- TOC entry 3717 (class 2606 OID 17512)
-- Name: season unique_season_per_title; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.season
    ADD CONSTRAINT unique_season_per_title UNIQUE (title_id, season_number);


--
-- TOC entry 3665 (class 2606 OID 17315)
-- Name: title unique_title; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title
    ADD CONSTRAINT unique_title UNIQUE (title_ru, title_origin, release_year);


--
-- TOC entry 3709 (class 2606 OID 17459)
-- Name: title_person unique_title_person_prof; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_person
    ADD CONSTRAINT unique_title_person_prof UNIQUE (title_id, person_id, profession_id);


--
-- TOC entry 3739 (class 2606 OID 17616)
-- Name: user_collection user_collection_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_collection
    ADD CONSTRAINT user_collection_pkey PRIMARY KEY (collection_id);


--
-- TOC entry 3747 (class 2606 OID 17668)
-- Name: watch_progress watch_progress_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_progress
    ADD CONSTRAINT watch_progress_pkey PRIMARY KEY (progress_id);


--
-- TOC entry 3753 (class 2606 OID 16581)
-- Name: transaction_logs fk_tx_wallet; Type: FK CONSTRAINT; Schema: analytics_priv; Owner: postgres
--

ALTER TABLE ONLY analytics_priv.transaction_logs
    ADD CONSTRAINT fk_tx_wallet FOREIGN KEY (wallet_id) REFERENCES billing_secure.wallets(wallet_id);


--
-- TOC entry 3752 (class 2606 OID 16567)
-- Name: wallets fk_wallet_owner; Type: FK CONSTRAINT; Schema: billing_secure; Owner: postgres
--

ALTER TABLE ONLY billing_secure.wallets
    ADD CONSTRAINT fk_wallet_owner FOREIGN KEY (owner_id) REFERENCES core_data.users(user_id) ON DELETE CASCADE;


--
-- TOC entry 3760 (class 2606 OID 17402)
-- Name: award_category award_category_award_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award_category
    ADD CONSTRAINT award_category_award_id_fkey FOREIGN KEY (award_id) REFERENCES public.award(award_id) ON DELETE CASCADE;


--
-- TOC entry 3761 (class 2606 OID 17407)
-- Name: award_category award_category_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award_category
    ADD CONSTRAINT award_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(category_id) ON DELETE CASCADE;


--
-- TOC entry 3759 (class 2606 OID 17382)
-- Name: award award_country_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_country_id_fkey FOREIGN KEY (country_id) REFERENCES public.country(country_id) ON DELETE CASCADE;


--
-- TOC entry 3778 (class 2606 OID 17654)
-- Name: content_info content_info_episode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.content_info
    ADD CONSTRAINT content_info_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episode(episode_id) ON DELETE CASCADE;


--
-- TOC entry 3779 (class 2606 OID 17649)
-- Name: content_info content_info_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.content_info
    ADD CONSTRAINT content_info_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3770 (class 2606 OID 17533)
-- Name: episode episode_season_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.episode
    ADD CONSTRAINT episode_season_id_fkey FOREIGN KEY (season_id) REFERENCES public.season(season_id) ON DELETE CASCADE;


--
-- TOC entry 3774 (class 2606 OID 17593)
-- Name: episode_tag episode_tag_episode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.episode_tag
    ADD CONSTRAINT episode_tag_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episode(episode_id) ON DELETE CASCADE;


--
-- TOC entry 3775 (class 2606 OID 17598)
-- Name: episode_tag episode_tag_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.episode_tag
    ADD CONSTRAINT episode_tag_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tag(tag_id) ON DELETE CASCADE;


--
-- TOC entry 3762 (class 2606 OID 17445)
-- Name: profession profession_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profession
    ADD CONSTRAINT profession_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(department_id) ON DELETE CASCADE;


--
-- TOC entry 3754 (class 2606 OID 17288)
-- Name: profile profile_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.account(account_id) ON DELETE CASCADE;


--
-- TOC entry 3769 (class 2606 OID 17513)
-- Name: season season_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.season
    ADD CONSTRAINT season_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3771 (class 2606 OID 17563)
-- Name: tag tag_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.tag_category(category_id);


--
-- TOC entry 3766 (class 2606 OID 17494)
-- Name: title_award title_award_award_id_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_award
    ADD CONSTRAINT title_award_award_id_category_id_fkey FOREIGN KEY (award_id, category_id) REFERENCES public.award_category(award_id, category_id);


--
-- TOC entry 3767 (class 2606 OID 17489)
-- Name: title_award title_award_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_award
    ADD CONSTRAINT title_award_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON DELETE CASCADE;


--
-- TOC entry 3768 (class 2606 OID 17484)
-- Name: title_award title_award_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_award
    ADD CONSTRAINT title_award_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3757 (class 2606 OID 17365)
-- Name: title_country title_country_country_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_country
    ADD CONSTRAINT title_country_country_id_fkey FOREIGN KEY (country_id) REFERENCES public.country(country_id) ON DELETE CASCADE;


--
-- TOC entry 3758 (class 2606 OID 17360)
-- Name: title_country title_country_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_country
    ADD CONSTRAINT title_country_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3755 (class 2606 OID 17338)
-- Name: title_genre title_genre_genre_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_genre
    ADD CONSTRAINT title_genre_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genre(genre_id) ON DELETE CASCADE;


--
-- TOC entry 3756 (class 2606 OID 17333)
-- Name: title_genre title_genre_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_genre
    ADD CONSTRAINT title_genre_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3763 (class 2606 OID 17465)
-- Name: title_person title_person_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_person
    ADD CONSTRAINT title_person_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON DELETE CASCADE;


--
-- TOC entry 3764 (class 2606 OID 17470)
-- Name: title_person title_person_profession_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_person
    ADD CONSTRAINT title_person_profession_id_fkey FOREIGN KEY (profession_id) REFERENCES public.profession(profession_id) ON DELETE RESTRICT;


--
-- TOC entry 3765 (class 2606 OID 17460)
-- Name: title_person title_person_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_person
    ADD CONSTRAINT title_person_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3772 (class 2606 OID 17581)
-- Name: title_tag title_tag_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_tag
    ADD CONSTRAINT title_tag_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tag(tag_id) ON DELETE CASCADE;


--
-- TOC entry 3773 (class 2606 OID 17576)
-- Name: title_tag title_tag_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.title_tag
    ADD CONSTRAINT title_tag_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3783 (class 2606 OID 17722)
-- Name: trivia_link trivia_link_episode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.trivia_link
    ADD CONSTRAINT trivia_link_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episode(episode_id) ON DELETE CASCADE;


--
-- TOC entry 3784 (class 2606 OID 17717)
-- Name: trivia_link trivia_link_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.trivia_link
    ADD CONSTRAINT trivia_link_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON DELETE CASCADE;


--
-- TOC entry 3785 (class 2606 OID 17712)
-- Name: trivia_link trivia_link_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.trivia_link
    ADD CONSTRAINT trivia_link_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3786 (class 2606 OID 17707)
-- Name: trivia_link trivia_link_trivia_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.trivia_link
    ADD CONSTRAINT trivia_link_trivia_id_fkey FOREIGN KEY (trivia_id) REFERENCES public.trivia(trivia_id) ON DELETE CASCADE;


--
-- TOC entry 3776 (class 2606 OID 17619)
-- Name: user_collection user_collection_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_collection
    ADD CONSTRAINT user_collection_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profile(profile_id) ON DELETE CASCADE;


--
-- TOC entry 3777 (class 2606 OID 17624)
-- Name: user_collection user_collection_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_collection
    ADD CONSTRAINT user_collection_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3780 (class 2606 OID 17681)
-- Name: watch_progress watch_progress_episode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_progress
    ADD CONSTRAINT watch_progress_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episode(episode_id) ON DELETE CASCADE;


--
-- TOC entry 3781 (class 2606 OID 17671)
-- Name: watch_progress watch_progress_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_progress
    ADD CONSTRAINT watch_progress_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.profile(profile_id) ON DELETE CASCADE;


--
-- TOC entry 3782 (class 2606 OID 17676)
-- Name: watch_progress watch_progress_title_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_progress
    ADD CONSTRAINT watch_progress_title_id_fkey FOREIGN KEY (title_id) REFERENCES public.title(title_id) ON DELETE CASCADE;


--
-- TOC entry 3990 (class 0 OID 0)
-- Dependencies: 11
-- Name: SCHEMA analytics_priv; Type: ACL; Schema: -; Owner: schema_manager
--

GRANT USAGE ON SCHEMA analytics_priv TO app_readonly;
GRANT USAGE ON SCHEMA analytics_priv TO app_writer;


--
-- TOC entry 3991 (class 0 OID 0)
-- Dependencies: 10
-- Name: SCHEMA billing_secure; Type: ACL; Schema: -; Owner: schema_manager
--

GRANT USAGE ON SCHEMA billing_secure TO app_readonly;
GRANT USAGE ON SCHEMA billing_secure TO app_writer;


--
-- TOC entry 3992 (class 0 OID 0)
-- Dependencies: 9
-- Name: SCHEMA core_data; Type: ACL; Schema: -; Owner: schema_manager
--

GRANT USAGE ON SCHEMA core_data TO app_readonly;


--
-- TOC entry 3996 (class 0 OID 0)
-- Dependencies: 227
-- Name: TABLE transaction_logs; Type: ACL; Schema: analytics_priv; Owner: postgres
--

GRANT SELECT ON TABLE analytics_priv.transaction_logs TO app_readonly;
GRANT SELECT,INSERT ON TABLE analytics_priv.transaction_logs TO app_writer;


--
-- TOC entry 3998 (class 0 OID 0)
-- Dependencies: 225
-- Name: TABLE wallets; Type: ACL; Schema: billing_secure; Owner: postgres
--

GRANT SELECT ON TABLE billing_secure.wallets TO app_readonly;
GRANT SELECT,UPDATE ON TABLE billing_secure.wallets TO app_writer;


--
-- TOC entry 4000 (class 0 OID 0)
-- Dependencies: 223
-- Name: TABLE users; Type: ACL; Schema: core_data; Owner: postgres
--

GRANT SELECT ON TABLE core_data.users TO app_readonly;


--
-- TOC entry 4001 (class 0 OID 0)
-- Dependencies: 228
-- Name: TABLE v_rich_users_report; Type: ACL; Schema: core_data; Owner: postgres
--

GRANT SELECT ON TABLE core_data.v_rich_users_report TO app_readonly;


-- Completed on 2026-04-06 19:56:19

--
-- PostgreSQL database dump complete
--

