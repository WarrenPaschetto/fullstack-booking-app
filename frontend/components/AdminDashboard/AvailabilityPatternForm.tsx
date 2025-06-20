import React from 'react';
import DatePicker from 'react-datepicker';

interface Props {
    weekValue: number
    setWeekValue: (d: number) => void
    startTimeValue: Date | null
    setStartTimeValue: (d: Date | null) => void
    endTimeValue: Date | null
    setEndTimeValue: (d: Date | null) => void
    onSubmit: React.FormEventHandler<HTMLFormElement>
    errorMsg?: string
}

export default function AvailabilityPatternForm({
    weekValue,
    setWeekValue,
    startTimeValue,
    setStartTimeValue,
    endTimeValue,
    setEndTimeValue,
    onSubmit,
    errorMsg,
}: Props) {
    return (
        <form onSubmit={onSubmit}>
            {/* Heading */}
            <h2 className="text-3xl text-blue-50 font-bold text-center">
                Create Availability Pattern
            </h2>

            <div className="space-y-4">
                <label className="block">
                    <span className="block mb-1 text-lg text-blue-50 font-medium">
                        Weekday
                    </span>
                    <select
                        required
                        value={weekValue}
                        onChange={(e) => setWeekValue(Number(e.target.value))}
                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                    >
                        <option value="" disabled>
                            Select Weekday
                        </option>
                        <option value={0}>Sunday</option>
                        <option value={1}>Monday</option>
                        <option value={2}>Tuesday</option>
                        <option value={3}>Wednesday</option>
                        <option value={4}>Thursday</option>
                        <option value={5}>Friday</option>
                        <option value={6}>Saturday</option>
                    </select>
                </label>
                <label className="block mb-1 text-lg font-medium text-blue-50">
                    <span className="block">
                        Start Time:
                    </span>
                    <DatePicker
                        selected={startTimeValue}
                        locale="en-US"
                        onChange={(date) => setStartTimeValue(date)}
                        showTimeSelect
                        timeIntervals={30}
                        dateFormat="Pp"
                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                    />
                </label>
                <label className="block mb-1 text-lg font-medium text-blue-50">
                    <span className="block">
                        End Time:
                    </span>
                    <DatePicker
                        selected={endTimeValue}
                        locale="en-US"
                        onChange={(date) => setEndTimeValue(date)}
                        showTimeSelect
                        timeIntervals={30}
                        dateFormat="Pp"
                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                    />
                </label>
            </div>

            {errorMsg && (
                <p className="text-red-300 text-sm text-center">{errorMsg}</p>
            )}

            <button
                type="submit"
                className="w-full text-blue-700 bg-blue-200 bg-opacity-20 hover:bg-opacity-30 py-2 
          rounded-lg font-semibold transition disabled:opacity-50 hover:bg-blue-400"
            >
                Submit Changes
            </button>
        </form>
    )
}
